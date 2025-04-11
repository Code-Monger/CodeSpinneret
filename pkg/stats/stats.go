package stats

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ToolStats represents statistics for a single tool
type ToolStats struct {
	Name                 string        `json:"name"`
	CallCount            int           `json:"call_count"`
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	AverageExecutionTime time.Duration `json:"average_execution_time"`
	InputTokens          int           `json:"input_tokens"`
	OutputTokens         int           `json:"output_tokens"`
	TokensSaved          int           `json:"tokens_saved"`
	LastUsed             time.Time     `json:"last_used"`
}

// SessionStats represents statistics for the current session
type SessionStats struct {
	StartTime time.Time             `json:"start_time"`
	Tools     map[string]*ToolStats `json:"tools"`
}

// PersistentStats represents statistics persisted across all sessions
type PersistentStats struct {
	FirstRecorded time.Time             `json:"first_recorded"`
	LastUpdated   time.Time             `json:"last_updated"`
	Tools         map[string]*ToolStats `json:"tools"`
}

// StatsManager manages tool usage statistics
type StatsManager struct {
	sessionStats    *SessionStats
	persistentStats *PersistentStats
	statsFilePath   string
	mutex           sync.RWMutex
}

// NewStatsManager creates a new StatsManager
func NewStatsManager(statsFilePath string) (*StatsManager, error) {
	// Create a new StatsManager
	manager := &StatsManager{
		sessionStats: &SessionStats{
			StartTime: time.Now(),
			Tools:     make(map[string]*ToolStats),
		},
		persistentStats: &PersistentStats{
			FirstRecorded: time.Now(),
			LastUpdated:   time.Now(),
			Tools:         make(map[string]*ToolStats),
		},
		statsFilePath: statsFilePath,
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(statsFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for stats file: %v", err)
	}

	// Load persistent stats if they exist
	if _, err := os.Stat(statsFilePath); err == nil {
		data, err := ioutil.ReadFile(statsFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read stats file: %v", err)
		}

		if err := json.Unmarshal(data, &manager.persistentStats); err != nil {
			return nil, fmt.Errorf("failed to parse stats file: %v", err)
		}
	}

	return manager, nil
}

// RecordToolUsage records statistics for a tool usage
func (m *StatsManager) RecordToolUsage(toolName string, executionTime time.Duration, inputTokens, outputTokens int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Update session stats
	sessionTool, ok := m.sessionStats.Tools[toolName]
	if !ok {
		sessionTool = &ToolStats{
			Name:     toolName,
			LastUsed: time.Now(),
		}
		m.sessionStats.Tools[toolName] = sessionTool
	}

	sessionTool.CallCount++
	sessionTool.TotalExecutionTime += executionTime
	sessionTool.AverageExecutionTime = sessionTool.TotalExecutionTime / time.Duration(sessionTool.CallCount)
	sessionTool.InputTokens += inputTokens
	sessionTool.OutputTokens += outputTokens
	sessionTool.TokensSaved += estimateTokensSaved(toolName, inputTokens, outputTokens)
	sessionTool.LastUsed = time.Now()

	// Update persistent stats
	persistentTool, ok := m.persistentStats.Tools[toolName]
	if !ok {
		persistentTool = &ToolStats{
			Name:     toolName,
			LastUsed: time.Now(),
		}
		m.persistentStats.Tools[toolName] = persistentTool
	}

	persistentTool.CallCount++
	persistentTool.TotalExecutionTime += executionTime
	persistentTool.AverageExecutionTime = persistentTool.TotalExecutionTime / time.Duration(persistentTool.CallCount)
	persistentTool.InputTokens += inputTokens
	persistentTool.OutputTokens += outputTokens
	persistentTool.TokensSaved += estimateTokensSaved(toolName, inputTokens, outputTokens)
	persistentTool.LastUsed = time.Now()
	m.persistentStats.LastUpdated = time.Now()

	// Save persistent stats to file
	return m.savePersistentStats()
}

// GetSessionStats returns statistics for the current session
func (m *StatsManager) GetSessionStats() *SessionStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a deep copy to avoid race conditions
	stats := &SessionStats{
		StartTime: m.sessionStats.StartTime,
		Tools:     make(map[string]*ToolStats),
	}

	for name, tool := range m.sessionStats.Tools {
		toolCopy := *tool
		stats.Tools[name] = &toolCopy
	}

	return stats
}

// GetPersistentStats returns statistics persisted across all sessions
func (m *StatsManager) GetPersistentStats() *PersistentStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Create a deep copy to avoid race conditions
	stats := &PersistentStats{
		FirstRecorded: m.persistentStats.FirstRecorded,
		LastUpdated:   m.persistentStats.LastUpdated,
		Tools:         make(map[string]*ToolStats),
	}

	for name, tool := range m.persistentStats.Tools {
		toolCopy := *tool
		stats.Tools[name] = &toolCopy
	}

	return stats
}

// savePersistentStats saves persistent stats to file
func (m *StatsManager) savePersistentStats() error {
	data, err := json.MarshalIndent(m.persistentStats, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal stats: %v", err)
	}

	if err := ioutil.WriteFile(m.statsFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write stats file: %v", err)
	}

	return nil
}

// estimateTokensSaved estimates the number of tokens saved by using a tool
func estimateTokensSaved(toolName string, inputTokens, outputTokens int) int {
	// This is a simple estimation based on the tool type
	// In a real implementation, this would be more sophisticated
	switch toolName {
	case "calculator":
		// Simple operations save minimal tokens
		return inputTokens / 2
	case "filesearch":
		// File search can save a lot of tokens by avoiding listing file contents
		return outputTokens * 3
	case "cmdexec":
		// Command execution can save a lot of tokens by avoiding manual steps
		return outputTokens * 4
	case "searchreplace":
		// Search and replace can save a lot of tokens by avoiding manual edits
		return outputTokens * 3
	case "screenshot":
		// Screenshots save a lot of tokens by avoiding descriptions
		return 1000 // Arbitrary large number for images
	case "websearch":
		// Web search saves tokens by providing external information
		return outputTokens * 5
	case "rag":
		// RAG saves tokens by providing context-aware responses
		return outputTokens * 6
	case "codeanalysis":
		// Code analysis saves tokens by providing insights
		return outputTokens * 4
	case "patch":
		// Patch saves tokens by applying changes efficiently
		return outputTokens * 3
	default:
		// Default estimation
		return outputTokens
	}
}

// FormatStats formats statistics as a string
func FormatStats(sessionStats *SessionStats, persistentStats *PersistentStats) string {
	result := "Tool Usage Statistics\n\n"

	// Session stats
	result += "Current Session Statistics:\n"
	result += fmt.Sprintf("Session started: %s\n", sessionStats.StartTime.Format(time.RFC3339))
	result += fmt.Sprintf("Session duration: %s\n\n", time.Since(sessionStats.StartTime).Round(time.Second))

	if len(sessionStats.Tools) > 0 {
		result += "Tool                  | Calls | Avg Time  | Total Time | Tokens Saved\n"
		result += "----------------------|-------|-----------|------------|-------------\n"

		for _, tool := range sessionStats.Tools {
			result += fmt.Sprintf("%-22s| %5d | %9s | %10s | %12d\n",
				tool.Name,
				tool.CallCount,
				tool.AverageExecutionTime.Round(time.Millisecond).String(),
				tool.TotalExecutionTime.Round(time.Millisecond).String(),
				tool.TokensSaved)
		}
	} else {
		result += "No tools used in this session.\n"
	}

	// Persistent stats
	result += "\nAll-Time Statistics:\n"
	result += fmt.Sprintf("First recorded: %s\n", persistentStats.FirstRecorded.Format(time.RFC3339))
	result += fmt.Sprintf("Last updated: %s\n\n", persistentStats.LastUpdated.Format(time.RFC3339))

	if len(persistentStats.Tools) > 0 {
		result += "Tool                  | Calls | Avg Time  | Total Time | Tokens Saved\n"
		result += "----------------------|-------|-----------|------------|-------------\n"

		for _, tool := range persistentStats.Tools {
			result += fmt.Sprintf("%-22s| %5d | %9s | %10s | %12d\n",
				tool.Name,
				tool.CallCount,
				tool.AverageExecutionTime.Round(time.Millisecond).String(),
				tool.TotalExecutionTime.Round(time.Millisecond).String(),
				tool.TokensSaved)
		}
	} else {
		result += "No tools used across all sessions.\n"
	}

	return result
}
