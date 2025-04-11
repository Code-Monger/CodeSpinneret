package tools

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestStats tests the stats tool
func TestStats(ctx context.Context, c client.MCPClient) error {
	log.Printf("Running stats test")

	callReq := mcp.CallToolRequest{}
	callReq.Params.Name = "stats"
	callReq.Params.Arguments = map[string]interface{}{}

	result, err := c.CallTool(ctx, callReq)
	if err != nil {
		log.Printf("Failed to call stats: %v", err)
		return err
	}

	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			log.Printf("Stats result:\n%s", textContent.Text)
		}
	}

	return nil
}
