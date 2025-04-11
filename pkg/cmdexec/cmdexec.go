package cmdexec

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleCommandExecution is the handler function for the command execution tool
func HandleCommandExecution(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract command
	command, ok := arguments["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command must be a string")
	}

	// Extract working directory (optional)
	workingDir, _ := arguments["working_directory"].(string)

	// Extract timeout (optional)
	timeoutSec, ok := arguments["timeout"].(float64)
	if !ok {
		// Default timeout: 30 seconds
		timeoutSec = 30
	}

	// Create a context with timeout
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	// Prepare the command based on the OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(execCtx, "sh", "-c", command)
	}

	// Set working directory if provided
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Prepare the result
	var resultText string
	if err != nil {
		// Include the error in the result
		resultText = fmt.Sprintf("Command execution failed: %v\n\n", err)
	} else {
		resultText = "Command executed successfully\n\n"
	}

	// Add stdout and stderr to the result
	if stdout.Len() > 0 {
		resultText += fmt.Sprintf("Standard Output:\n%s\n", stdout.String())
	}
	if stderr.Len() > 0 {
		resultText += fmt.Sprintf("Standard Error:\n%s\n", stderr.String())
	}

	// Add command information
	resultText += fmt.Sprintf("\nCommand: %s\n", command)
	if workingDir != "" {
		resultText += fmt.Sprintf("Working Directory: %s\n", workingDir)
	}

	// Add exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1 // Unknown error
		}
	}
	resultText += fmt.Sprintf("Exit Code: %d\n", exitCode)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// RegisterCommandExecution registers the command execution tool with the MCP server
func RegisterCommandExecution(mcpServer *server.MCPServer) {
	mcpServer.AddTool(mcp.NewTool("cmdexec",
		mcp.WithDescription("Execute commands on the system, such as running scripts, compiling code, or starting applications"),
		mcp.WithString("command",
			mcp.Description("The command to execute"),
			mcp.Required(),
		),
		mcp.WithString("working_directory",
			mcp.Description("The working directory for the command (optional)"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Timeout in seconds (default: 30)"),
		),
	), HandleCommandExecution)
}
