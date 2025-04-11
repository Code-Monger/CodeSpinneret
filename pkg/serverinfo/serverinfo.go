package serverinfo

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleServerInfo is the handler function for the server info resource
func HandleServerInfo(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	info := map[string]interface{}{
		"timestamp":      time.Now().Format(time.RFC3339),
		"go_version":     runtime.Version(),
		"os":             runtime.GOOS,
		"architecture":   runtime.GOARCH,
		"cpu_cores":      runtime.NumCPU(),
		"goroutines":     runtime.NumGoroutine(),
		"memory_stats":   getMemoryStats(),
		"uptime_seconds": getUptime(),
	}

	// Convert to JSON string
	infoStr := fmt.Sprintf("Server Information:\n\n")
	for k, v := range info {
		infoStr += fmt.Sprintf("%s: %v\n", k, v)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/plain",
			Text:     infoStr,
		},
	}, nil
}

// RegisterServerInfo registers the server info resource with the MCP server
func RegisterServerInfo(mcpServer *server.MCPServer) {
	mcpServer.AddResource(
		mcp.NewResource(
			"server://info",
			"Server Information",
			mcp.WithMIMEType("text/plain"),
		),
		HandleServerInfo,
	)
}

// getMemoryStats returns memory statistics
func getMemoryStats() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return map[string]interface{}{
		"alloc_mb":       float64(memStats.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(memStats.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(memStats.Sys) / 1024 / 1024,
		"num_gc":         memStats.NumGC,
	}
}

// startTime is used to calculate uptime
var startTime = time.Now()

// getUptime returns the server uptime in seconds
func getUptime() float64 {
	return time.Since(startTime).Seconds()
}
