package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/picoclaw/pkg/bus"
	"github.com/sipeed/picoclaw/pkg/providers"
)

type SubagentTask struct {
	ID            string
	Task          string
	Label         string
	OriginChannel string
	OriginChatID  string
	Status        string
	Result        string
	Created       int64
}

type SubagentManager struct {
	tasks     map[string]*SubagentTask
	mu        sync.RWMutex
	provider  providers.LLMProvider
	bus       *bus.MessageBus
	workspace string
	registry  *ToolRegistry
}

func NewSubagentManager(provider providers.LLMProvider, workspace string, bus *bus.MessageBus, registry *ToolRegistry) *SubagentManager {
	return &SubagentManager{
		tasks:     make(map[string]*SubagentTask),
		provider:  provider,
		bus:       bus,
		workspace: workspace,
		registry:  registry,
	}
}

func (sm *SubagentManager) Spawn(ctx context.Context, task, label, originChannel, originChatID string) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	taskID := fmt.Sprintf("subagent-%d", time.Now().UnixNano())

	subagentTask := &SubagentTask{
		ID:            taskID,
		Task:          task,
		Label:         label,
		OriginChannel: originChannel,
		OriginChatID:  originChatID,
		Status:        "running",
		Created:       time.Now().UnixMilli(),
	}
	sm.tasks[taskID] = subagentTask

	go sm.runTask(ctx, subagentTask)

	if label != "" {
		return fmt.Sprintf("Spawned subagent '%s' for task: %s", label, task), nil
	}
	return fmt.Sprintf("Spawned subagent for task: %s", task), nil
}

func (sm *SubagentManager) runTask(ctx context.Context, task *SubagentTask) {
	task.Status = "running"
	task.Created = time.Now().UnixMilli()

	messages := []providers.Message{
		{
			Role:    "system",
			Content: "You are a subagent. Complete the given task independently and report the result. You have access to tools. Use them to gather information and perform actions.",
		},
		{
			Role:    "user",
			Content: task.Task,
		},
	}

	maxIterations := 10
	iteration := 0
	var finalResult string

	for iteration < maxIterations {
		iteration++
		
		var providerToolDefs []providers.ToolDefinition
		if sm.registry != nil {
			toolDefs := sm.registry.GetDefinitions()
			for _, td := range toolDefs {
				funcMap := td["function"].(map[string]interface{})
				if funcMap["name"].(string) == "spawn" {
					continue
				}
				providerToolDefs = append(providerToolDefs, providers.ToolDefinition{
					Type: td["type"].(string),
					Function: providers.ToolFunctionDefinition{
						Name:        funcMap["name"].(string),
						Description: funcMap["description"].(string),
						Parameters:  funcMap["parameters"].(map[string]interface{}),
					},
				})
			}
		}

		var response *providers.LLMResponse
		var err error
		maxRetries := 3
		backoff := 2 * time.Second

		for attempt := 1; attempt <= maxRetries; attempt++ {
			response, err = sm.provider.Chat(ctx, messages, providerToolDefs, sm.provider.GetDefaultModel(), map[string]interface{}{
				"max_tokens": 4096,
			})
			if err == nil {
				break
			}
			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					finalResult = fmt.Sprintf("Error: %v", ctx.Err())
					break
				case <-time.After(backoff):
					backoff *= 2
				}
			}
		}

		if err != nil {
			finalResult = fmt.Sprintf("Error after %d attempts: %v", maxRetries, err)
			break
		}

		if len(response.ToolCalls) == 0 {
			finalResult = response.Content
			break
		}

		assistantMsg := providers.Message{
			Role:    "assistant",
			Content: response.Content,
		}
		for _, tc := range response.ToolCalls {
			argsJSON, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: &providers.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argsJSON),
				},
			})
		}
		messages = append(messages, assistantMsg)

		for _, tc := range response.ToolCalls {
			result, err := sm.registry.ExecuteWithContext(ctx, tc.Name, tc.Arguments, task.OriginChannel, task.OriginChatID)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}
			messages = append(messages, providers.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	if iteration >= maxIterations && finalResult == "" {
		finalResult = "Error: Max iterations reached without a final answer."
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if strings.HasPrefix(finalResult, "Error:") {
		task.Status = "failed"
	} else {
		task.Status = "completed"
	}
	task.Result = finalResult

	// Send announce message back to main agent
	if sm.bus != nil {
		announceContent := fmt.Sprintf("Task '%s' completed.\n\nResult:\n%s", task.Label, task.Result)
		sm.bus.PublishInbound(bus.InboundMessage{
			Channel:  "system",
			SenderID: fmt.Sprintf("subagent:%s", task.ID),
			// Format: "original_channel:original_chat_id" for routing back
			ChatID:  fmt.Sprintf("%s:%s", task.OriginChannel, task.OriginChatID),
			Content: announceContent,
		})
	}
}

func (sm *SubagentManager) GetTask(taskID string) (*SubagentTask, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	task, ok := sm.tasks[taskID]
	return task, ok
}

func (sm *SubagentManager) ListTasks() []*SubagentTask {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tasks := make([]*SubagentTask, 0, len(sm.tasks))
	for _, task := range sm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
