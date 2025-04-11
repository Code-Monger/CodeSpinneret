package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFileSearch tests the file search tool with various search criteria
func TestFileSearch(ctx context.Context, c client.MCPClient) error {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Basic directory listing",
			arguments: map[string]interface{}{
				"directory": ".",
				"pattern":   "*.go",
				"recursive": false,
			},
		},
		{
			name: "Recursive search",
			arguments: map[string]interface{}{
				"directory": ".",
				"pattern":   "*.go",
				"recursive": true,
			},
		},
		{
			name: "Content search",
			arguments: map[string]interface{}{
				"directory":       ".",
				"pattern":         "*.go",
				"recursive":       true,
				"content_pattern": "func.*\\(",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running file search test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "filesearch"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call filesearch: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("File search result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
