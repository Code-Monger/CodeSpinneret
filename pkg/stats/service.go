package stats

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	// Global stats manager instance
	globalStatsManager *StatsManager
)

// InitStatsManager initializes the global stats manager
func InitStatsManager(dataDir string) error {
	statsFilePath := filepath.Join(dataDir, "stats.json")
	var err error
	globalStatsManager, err = NewStatsManager(statsFilePath)
	return err
}

// GetStatsManager returns the global stats manager
func GetStatsManager() *StatsManager {
	return globalStatsManager
}

// HandleGetStats handles requests to get tool usage statistics
func HandleGetStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("[Stats] Received request to get stats")

	if globalStatsManager == nil {
		log.Printf("[Stats] Error: stats manager not initialized")
		return nil, fmt.Errorf("stats manager not initialized")
	}

	// Get the stats
	sessionStats := globalStatsManager.GetSessionStats()
	persistentStats := globalStatsManager.GetPersistentStats()

	// Format the stats
	statsText := FormatStats(sessionStats, persistentStats)

	log.Printf("[Stats] Returning stats information")

	// Return the result
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: statsText,
			},
		},
	}, nil
}

// RecordToolUsage records statistics for a tool usage
func RecordToolUsage(toolName string, startTime time.Time, result *mcp.CallToolResult) {
	if globalStatsManager == nil {
		log.Printf("[Stats] Warning: stats manager not initialized, cannot record tool usage")
		return
	}

	// Record the execution time
	executionTime := time.Since(startTime)

	// Estimate input tokens (this is a simple estimation)
	inputTokens := 100 // Default value

	// Estimate output tokens (this is a simple estimation)
	outputTokens := estimateOutputTokens(result)

	log.Printf("[Stats] Recording usage for tool '%s': execution time=%v, input tokens=%d, output tokens=%d",
		toolName, executionTime, inputTokens, outputTokens)

	// Record the usage
	if err := globalStatsManager.RecordToolUsage(toolName, executionTime, inputTokens, outputTokens); err != nil {
		// Log the error but don't fail the request
		log.Printf("[Stats] Failed to record tool usage: %v", err)
	}
}

// WrapHandler wraps a tool handler with stats tracking
func WrapHandler(toolName string, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Record the start time
		startTime := time.Now()

		log.Printf("[Stats] Starting execution of tool '%s'", toolName)

		// Call the original handler
		result, err := handler(ctx, request)
		if err != nil {
			log.Printf("[Stats] Error executing tool '%s': %v", toolName, err)
			return nil, err
		}

		// Record the usage
		RecordToolUsage(toolName, startTime, result)

		return result, nil
	}
}

// HandleClientDisconnect handles a client disconnection
func HandleClientDisconnect(sessionID string) {
	if globalStatsManager == nil {
		log.Printf("[Stats] Warning: stats manager not initialized, cannot handle client disconnect")
		return
	}

	log.Printf("[Stats] Client disconnected: %s", sessionID)

	// Get the session stats
	sessionStats := globalStatsManager.GetSessionStats()
	persistentStats := globalStatsManager.GetPersistentStats()

	// Format and print the stats
	statsText := FormatStats(sessionStats, persistentStats)
	log.Printf("[Stats] Session statistics for client %s:\n%s", sessionID, statsText)

	// In a real implementation, we would remove the session metrics from RAM here
	// For now, we'll just log that we would do this
	log.Printf("[Stats] Removing session metrics for client %s from RAM", sessionID)

	// Reset session stats
	globalStatsManager.ResetSessionStats()
}

// ResetSessionStats resets the session statistics
func (m *StatsManager) ResetSessionStats() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessionStats = &SessionStats{
		StartTime: time.Now(),
		Tools:     make(map[string]*ToolStats),
	}

	log.Printf("[Stats] Session statistics reset")
}

// estimateInputTokens estimates the number of tokens in the request
func estimateInputTokens(request mcp.CallToolRequest) int {
	// This is a simple estimation based on the request size
	// In a real implementation, this would be more sophisticated
	tokens := 0

	// Count the tool name
	tokens += len(request.Params.Name)

	// Count the arguments
	for key, value := range request.Params.Arguments {
		tokens += len(key)
		switch v := value.(type) {
		case string:
			tokens += len(v)
		case float64:
			tokens += 1
		case bool:
			tokens += 1
		case []interface{}:
			tokens += len(v)
		}
	}

	return tokens
}

// estimateOutputTokens estimates the number of tokens in the result
func estimateOutputTokens(result *mcp.CallToolResult) int {
	// This is a simple estimation based on the result size
	// In a real implementation, this would be more sophisticated
	tokens := 0

	for _, content := range result.Content {
		switch c := content.(type) {
		case mcp.TextContent:
			tokens += len(c.Text)
		case mcp.ImageContent:
			// Images are worth a lot of tokens
			tokens += 1000
		}
	}

	return tokens
}

// RegisterStats registers the stats tool with the MCP server
func RegisterStats(mcpServer *server.MCPServer, dataDir string) error {
	// Initialize the stats manager
	if err := InitStatsManager(dataDir); err != nil {
		return err
	}

	// Create the tool definition
	statsTool := mcp.NewTool("stats",
		mcp.WithDescription("Retrieves usage statistics for MCP tools"),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := WrapHandler("stats", HandleGetStats)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(statsTool, wrappedHandler)

	log.Printf("[Stats] Registered stats tool")

	return nil
}
