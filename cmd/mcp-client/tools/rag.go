package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestRAG tests the RAG tool with various operations
func TestRAG(ctx context.Context, c client.MCPClient) error {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Index repository",
			arguments: map[string]interface{}{
				"operation":     "index",
				"repo_path":     ".",
				"file_patterns": []interface{}{"*.go", "*.md"},
			},
		},
		{
			name: "Query repository",
			arguments: map[string]interface{}{
				"operation":   "query",
				"repo_path":   ".",
				"query":       "how to implement a service",
				"num_results": 3.0,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running RAG test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "rag"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call rag: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("RAG result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(1 * time.Second)
	}

	return nil
}
