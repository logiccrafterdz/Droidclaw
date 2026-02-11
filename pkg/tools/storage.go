package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StorageTool provides structured data storage within the workspace.
// All operations are sandboxed to the workspace/data/ directory for security.
type StorageTool struct {
	dataDir string
}

func NewStorageTool(workspace string) *StorageTool {
	dataDir := filepath.Join(workspace, "data")
	os.MkdirAll(dataDir, 0755)
	return &StorageTool{dataDir: dataDir}
}

func (t *StorageTool) Name() string {
	return "storage"
}

func (t *StorageTool) Description() string {
	return `Read, write, append, and list structured data files in the workspace data directory.
Actions:
- "write": Write JSON data to a file (creates directories automatically)
- "read": Read a file's contents
- "append": Append a JSON entry to an existing file (for logs/time-series)
- "list": List files in a directory
- "delete": Delete a file
All paths are relative to the workspace data/ directory. Example paths: "scans/2025-01-15.json", "reports/daily.json"`
}

func (t *StorageTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"write", "read", "append", "list", "delete"},
				"description": "Storage action to perform",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path relative to data/ directory (e.g., 'scans/latest.json')",
			},
			"data": map[string]interface{}{
				"type":        "string",
				"description": "Data to write or append (JSON string or plain text)",
			},
		},
		"required": []string{"action"},
	}
}

func (t *StorageTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "write":
		return t.writeData(args)
	case "read":
		return t.readData(args)
	case "append":
		return t.appendData(args)
	case "list":
		return t.listData(args)
	case "delete":
		return t.deleteData(args)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *StorageTool) resolvePath(relPath string) (string, error) {
	if relPath == "" {
		return "", fmt.Errorf("path is required")
	}

	// Security: prevent path traversal
	clean := filepath.Clean(relPath)
	if strings.Contains(clean, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}

	absPath := filepath.Join(t.dataDir, clean)

	// Verify it's still within dataDir
	absDataDir, _ := filepath.Abs(t.dataDir)
	absResolved, _ := filepath.Abs(absPath)
	if !strings.HasPrefix(absResolved, absDataDir) {
		return "", fmt.Errorf("path outside data directory")
	}

	return absPath, nil
}

func (t *StorageTool) writeData(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	data, _ := args["data"].(string)

	if path == "" {
		return "Error: path is required for write", nil
	}
	if data == "" {
		return "Error: data is required for write", nil
	}

	absPath, err := t.resolvePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// Create parent directories
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err), nil
	}

	if err := os.WriteFile(absPath, []byte(data), 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	return fmt.Sprintf("Written %d bytes to %s", len(data), path), nil
}

func (t *StorageTool) readData(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "Error: path is required for read", nil
	}

	absPath, err := t.resolvePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("File not found: %s", path), nil
		}
		return fmt.Sprintf("Error reading file: %v", err), nil
	}

	// Truncate very large files
	result := string(content)
	if len(result) > 50000 {
		result = result[:50000] + fmt.Sprintf("\n... (truncated, total %d bytes)", len(content))
	}

	return result, nil
}

func (t *StorageTool) appendData(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	data, _ := args["data"].(string)

	if path == "" {
		return "Error: path is required for append", nil
	}
	if data == "" {
		return "Error: data is required for append", nil
	}

	absPath, err := t.resolvePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	// Create parent directories
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Sprintf("Error creating directory: %v", err), nil
	}

	// For JSON files, try to append to a JSON array
	if strings.HasSuffix(path, ".json") {
		return t.appendJSON(absPath, data, path)
	}

	// For other files, simple line append with timestamp
	entry := fmt.Sprintf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), data)
	f, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Sprintf("Error opening file: %v", err), nil
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Sprintf("Error appending: %v", err), nil
	}

	return fmt.Sprintf("Appended to %s", path), nil
}

func (t *StorageTool) appendJSON(absPath, data, relPath string) (string, error) {
	// Read existing content
	var entries []json.RawMessage
	if existing, err := os.ReadFile(absPath); err == nil {
		// Try to parse as array
		if err := json.Unmarshal(existing, &entries); err != nil {
			// If not an array, wrap existing content as first entry
			entries = []json.RawMessage{existing}
		}
	}

	// Parse new data as JSON, wrap in raw message
	var newEntry json.RawMessage
	if err := json.Unmarshal([]byte(data), &newEntry); err != nil {
		// If not valid JSON, wrap as string
		wrapped, _ := json.Marshal(data)
		newEntry = wrapped
	}

	entries = append(entries, newEntry)

	// Write back
	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err), nil
	}

	if err := os.WriteFile(absPath, out, 0644); err != nil {
		return fmt.Sprintf("Error writing file: %v", err), nil
	}

	return fmt.Sprintf("Appended JSON entry to %s (total: %d entries)", relPath, len(entries)), nil
}

func (t *StorageTool) listData(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}

	absPath, err := t.resolvePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("Directory not found: %s", path), nil
		}
		return fmt.Sprintf("Error listing directory: %v", err), nil
	}

	type fileInfo struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Size    int64  `json:"size,omitempty"`
		ModTime string `json:"modified,omitempty"`
	}

	files := make([]fileInfo, 0, len(entries))
	for _, entry := range entries {
		fi := fileInfo{
			Name: entry.Name(),
			Type: "file",
		}
		if entry.IsDir() {
			fi.Type = "dir"
		}
		if info, err := entry.Info(); err == nil {
			fi.Size = info.Size()
			fi.ModTime = info.ModTime().Format("2006-01-02 15:04")
		}
		files = append(files, fi)
	}

	result := map[string]interface{}{
		"path":    path,
		"count":   len(files),
		"entries": files,
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

func (t *StorageTool) deleteData(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "Error: path is required for delete", nil
	}

	absPath, err := t.resolvePath(path)
	if err != nil {
		return fmt.Sprintf("Error: %v", err), nil
	}

	if err := os.Remove(absPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("File not found: %s", path), nil
		}
		return fmt.Sprintf("Error deleting file: %v", err), nil
	}

	return fmt.Sprintf("Deleted %s", path), nil
}
