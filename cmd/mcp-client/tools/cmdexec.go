package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestCommandExecution tests the command execution tool with various commands
func TestCommandExecution(ctx context.Context, c client.MCPClient) error {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Simple echo command",
			arguments: map[string]interface{}{
				"command": "echo Hello, World!",
				"timeout": 5.0,
			},
		},
		{
			name: "Directory listing",
			arguments: map[string]interface{}{
				"command": "dir",
				"timeout": 5.0,
			},
		},
		{
			name: "Current working directory",
			arguments: map[string]interface{}{
				"command": "cd",
				"timeout": 5.0,
			},
		},
		{
			name: "Test timeout functionality",
			arguments: map[string]interface{}{
				"command": "ping -n 10 127.0.0.1", // This command will take about 10 seconds
				"timeout": 2.0,                    // But we set timeout to 2 seconds
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running command execution test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "cmdexec"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call cmdexec: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Command execution result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}

	return nil
}
