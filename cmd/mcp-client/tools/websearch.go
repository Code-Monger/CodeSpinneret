package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestWebSearch tests the web search tool with various queries
func TestWebSearch(ctx context.Context, c client.MCPClient) error {
	// Define test cases - focusing on all three search engines with different queries
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		// DuckDuckGo tests
		{
			name: "DuckDuckGo - Programming search",
			arguments: map[string]interface{}{
				"query":       "golang programming",
				"num_results": 5.0,
				"engine":      "duckduckgo",
				"safe_search": true,
			},
		},
		{
			name: "DuckDuckGo - Technology search",
			arguments: map[string]interface{}{
				"query":       "artificial intelligence",
				"num_results": 5.0,
				"engine":      "duckduckgo",
				"safe_search": true,
			},
		},

		// Bing tests
		{
			name: "Bing - Programming search",
			arguments: map[string]interface{}{
				"query":       "golang programming",
				"num_results": 5.0,
				"engine":      "bing",
				"safe_search": true,
			},
		},
		{
			name: "Bing - Technology search",
			arguments: map[string]interface{}{
				"query":       "artificial intelligence",
				"num_results": 5.0,
				"engine":      "bing",
				"safe_search": true,
			},
		},

		// Google tests
		{
			name: "Google - Programming search",
			arguments: map[string]interface{}{
				"query":       "golang programming",
				"num_results": 5.0,
				"engine":      "google",
				"safe_search": true,
			},
		},
		{
			name: "Google - Technology search",
			arguments: map[string]interface{}{
				"query":       "artificial intelligence",
				"num_results": 5.0,
				"engine":      "google",
				"safe_search": true,
			},
		},
		{
			name: "Google - Science search",
			arguments: map[string]interface{}{
				"query":       "quantum physics",
				"num_results": 5.0,
				"engine":      "google",
				"safe_search": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running search test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "websearch"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call websearch: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Search result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(2 * time.Second)
	}

	return nil
}
