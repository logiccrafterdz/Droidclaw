// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MemoryStore manages persistent multi-tier neural memory for the agent.
// - Semantic Memory: memory/semantic.md (Rules, patterns, persistent facts)
// - Episodic Memory: memory/episodic/YYYYMMDD.md (Daily observations, chronological events)
// - Working Memory is handled by SessionManager.
type MemoryStore struct {
	workspace   string
	memoryDir   string
	semanticFile string
	episodicDir  string
}

// NewMemoryStore creates a new MemoryStore with the given workspace path.
// It ensures the memory directory exists.
func NewMemoryStore(workspace string) *MemoryStore {
	memoryDir := filepath.Join(workspace, "memory")
	semanticFile := filepath.Join(memoryDir, "semantic.md")
	episodicDir := filepath.Join(memoryDir, "episodic")

	// Ensure memory directories exist
	os.MkdirAll(memoryDir, 0755)
	os.MkdirAll(episodicDir, 0755)

	// Migrate old MEMORY.md to semantic.md if exists
	oldMemoryFile := filepath.Join(memoryDir, "MEMORY.md")
	if _, err := os.Stat(oldMemoryFile); err == nil {
		if _, err := os.Stat(semanticFile); os.IsNotExist(err) {
			os.Rename(oldMemoryFile, semanticFile)
		}
	}

	return &MemoryStore{
		workspace:   workspace,
		memoryDir:   memoryDir,
		semanticFile: semanticFile,
		episodicDir:  episodicDir,
	}
}

// getTodayFile returns the path to today's episodic memory file.
func (ms *MemoryStore) getTodayFile() string {
	today := time.Now().Format("20060102")      // YYYYMMDD
	filePath := filepath.Join(ms.episodicDir, today+".md")
	return filePath
}

// ReadSemantic reads the semantic memory (semantic.md).
func (ms *MemoryStore) ReadSemantic() string {
	if data, err := os.ReadFile(ms.semanticFile); err == nil {
		return string(data)
	}
	return ""
}

// WriteSemantic writes content to the semantic memory file.
func (ms *MemoryStore) WriteSemantic(content string) error {
	return os.WriteFile(ms.semanticFile, []byte(content), 0644)
}

// ReadToday reads today's daily note.
// Returns empty string if the file doesn't exist.
func (ms *MemoryStore) ReadToday() string {
	todayFile := ms.getTodayFile()
	if data, err := os.ReadFile(todayFile); err == nil {
		return string(data)
	}
	return ""
}

// AppendToday appends content to today's daily note.
// If the file doesn't exist, it creates a new file with a date header.
func (ms *MemoryStore) AppendToday(content string) error {
	todayFile := ms.getTodayFile()

	// Ensure month directory exists
	monthDir := filepath.Dir(todayFile)
	os.MkdirAll(monthDir, 0755)

	var existingContent string
	if data, err := os.ReadFile(todayFile); err == nil {
		existingContent = string(data)
	}

	var newContent string
	if existingContent == "" {
		// Add header for new day
		header := fmt.Sprintf("# %s\n\n", time.Now().Format("2006-01-02"))
		newContent = header + content
	} else {
		// Append to existing content
		newContent = existingContent + "\n" + content
	}

	return os.WriteFile(todayFile, []byte(newContent), 0644)
}

// GetRecentEpisodic returns episodic memories from the last N days.
func (ms *MemoryStore) GetRecentEpisodic(days int) string {
	var notes []string

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("20060102")      // YYYYMMDD
		filePath := filepath.Join(ms.episodicDir, dateStr+".md")

		if data, err := os.ReadFile(filePath); err == nil {
			notes = append(notes, string(data))
		}
	}

	if len(notes) == 0 {
		return ""
	}

	// Join with separator
	var result string
	for i, note := range notes {
		if i > 0 {
			result += "\n\n---\n\n"
		}
		result += note
	}
	return result
}

// GetMemoryContext returns formatted memory context for the agent prompt.
// Includes semantic memory and recent episodic memory.
func (ms *MemoryStore) GetMemoryContext() string {
	var parts []string

	// Semantic Memory (Long-term concepts/rules)
	semantic := ms.ReadSemantic()
	if semantic != "" {
		parts = append(parts, "## Semantic Memory (Rules, Patterns & Facts)\n\n"+semantic)
	}

	// Episodic Memory (Recent events)
	recentNotes := ms.GetRecentEpisodic(3)
	if recentNotes != "" {
		parts = append(parts, "## Episodic Memory (Recent Events)\n\n"+recentNotes)
	}

	if len(parts) == 0 {
		return ""
	}

	// Join parts with separator
	var result string
	for i, part := range parts {
		if i > 0 {
			result += "\n\n---\n\n"
		}
		result += part
	}
	return fmt.Sprintf("# Memory\n\n%s", result)
}
