package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestScreenshot tests the screenshot tool with various operations
func TestScreenshot(ctx context.Context, c client.MCPClient) error {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Full screen screenshot",
			arguments: map[string]interface{}{
				"area":   "full",
				"format": "png",
			},
		},
		{
			name: "Region screenshot",
			arguments: map[string]interface{}{
				"area":   "region",
				"x":      100.0,
				"y":      100.0,
				"width":  400.0,
				"height": 300.0,
				"format": "png",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running screenshot test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "screenshot"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call screenshot: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			for i, content := range result.Content {
				switch content.(type) {
				case mcp.TextContent:
					textContent := content.(mcp.TextContent)
					log.Printf("Screenshot result (text):\n%s", textContent.Text)
				case mcp.ImageContent:
					log.Printf("Screenshot result contains image data (content item %d)", i)
				default:
					log.Printf("Screenshot result contains unknown content type: %T", content)
				}
			}
		}

		// Add a small delay between tests
		time.Sleep(1 * time.Second)
	}

	return nil
}
