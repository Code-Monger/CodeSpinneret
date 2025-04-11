package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestWebFetch tests the web fetch tool
func TestWebFetch(ctx context.Context, c client.MCPClient) error {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Fetch with HTML stripped",
			arguments: map[string]interface{}{
				"url":            "https://example.com",
				"include_images": false,
				"strip_html":     true,
				"timeout":        10.0,
			},
		},
		{
			name: "Fetch HTML page",
			arguments: map[string]interface{}{
				"url":            "https://example.com",
				"include_images": false,
				"timeout":        10.0,
			},
		},
		{
			name: "Fetch with images",
			arguments: map[string]interface{}{
				"url":            "https://en.wikipedia.org/wiki/Main_Page",
				"include_images": true,
				"timeout":        15.0,
			},
		},
		{
			name: "Fetch with default scheme",
			arguments: map[string]interface{}{
				"url":            "golang.org",
				"include_images": false,
				"timeout":        10.0,
			},
		},
		{
			name: "Fetch with error (invalid URL)",
			arguments: map[string]interface{}{
				"url":            "https://this-domain-does-not-exist-12345.com",
				"include_images": false,
				"timeout":        5.0,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running web fetch test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "webfetch"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call webfetch: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				// Truncate the content for logging
				content := textContent.Text
				if len(content) > 500 {
					content = content[:500] + "... [truncated]"
				}
				log.Printf("Web fetch result:\n%s", content)
			}
		}

		// Add a small delay between tests
		time.Sleep(2 * time.Second)
	}

	return nil
}
