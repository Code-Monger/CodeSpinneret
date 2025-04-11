package stats

import (
	"context"
	"fmt"
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
	if globalStatsManager == nil {
		return nil, fmt.Errorf("stats manager not initialized")
	}

	// Get the stats
	sessionStats := globalStatsManager.GetSessionStats()
	persistentStats := globalStatsManager.GetPersistentStats()

	// Format the stats
	statsText := FormatStats(sessionStats, persistentStats)

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
		return
	}

	// Record the execution time
	executionTime := time.Since(startTime)

	// Estimate input tokens (this is a simple estimation)
	inputTokens := 100 // Default value

	// Estimate output tokens (this is a simple estimation)
	outputTokens := estimateOutputTokens(result)

	// Record the usage
	if err := globalStatsManager.RecordToolUsage(toolName, executionTime, inputTokens, outputTokens); err != nil {
		// Log the error but don't fail the request
		fmt.Printf("Failed to record tool usage: %v\n", err)
	}
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

	// Register the stats tool
	mcpServer.AddTool(mcp.NewTool("stats",
		mcp.WithDescription("Retrieves usage statistics for MCP tools"),
	), HandleGetStats)

	return nil
}
