package calculator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleCalculator is the handler function for the calculator tool
func HandleCalculator(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract operation
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	// Extract operands
	var a, b float64
	var err error

	// Handle different types that might come in for 'a'
	switch v := arguments["a"].(type) {
	case float64:
		a = v
	case int:
		a = float64(v)
	case string:
		a, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse 'a' as number: %v", err)
		}
	default:
		return nil, fmt.Errorf("'a' must be a number")
	}

	// Handle different types that might come in for 'b'
	switch v := arguments["b"].(type) {
	case float64:
		b = v
	case int:
		b = float64(v)
	case string:
		b, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse 'b' as number: %v", err)
		}
	default:
		return nil, fmt.Errorf("'b' must be a number")
	}

	// Perform the calculation
	var calcResult float64
	var resultText string

	switch operation {
	case "add":
		calcResult = a + b
		resultText = fmt.Sprintf("The sum of %g and %g is %g", a, b, calcResult)
	case "subtract":
		calcResult = a - b
		resultText = fmt.Sprintf("The difference of %g and %g is %g", a, b, calcResult)
	case "multiply":
		calcResult = a * b
		resultText = fmt.Sprintf("The product of %g and %g is %g", a, b, calcResult)
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero is not allowed")
		}
		calcResult = a / b
		resultText = fmt.Sprintf("The quotient of %g and %g is %g", a, b, calcResult)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// RegisterCalculator registers the calculator tool with the MCP server
func RegisterCalculator(mcpServer *server.MCPServer) {
	// Create the tool definition
	calculatorTool := mcp.NewTool("calculator",
		mcp.WithDescription("A simple calculator that can perform basic arithmetic operations"),
		mcp.WithString("operation",
			mcp.Description("The operation to perform (add, subtract, multiply, divide)"),
			mcp.Required(),
		),
		mcp.WithNumber("a",
			mcp.Description("First operand"),
			mcp.Required(),
		),
		mcp.WithNumber("b",
			mcp.Description("Second operand"),
			mcp.Required(),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("calculator", HandleCalculator)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(calculatorTool, wrappedHandler)
}
