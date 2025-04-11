package tools

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestCalculator tests the calculator tool with various operations
func TestCalculator(ctx context.Context, c client.MCPClient) error {
	operations := []struct {
		op string
		a  float64
		b  float64
	}{
		{"add", 5, 3},
		{"subtract", 10, 4},
		{"multiply", 6, 7},
		{"divide", 20, 5},
	}

	for _, op := range operations {
		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "calculator"
		callReq.Params.Arguments = map[string]interface{}{
			"operation": op.op,
			"a":         op.a,
			"b":         op.b,
		}

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call calculator with %s: %v", op.op, err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Calculator %s result: %s", op.op, textContent.Text)
			}
		}
	}

	return nil
}
