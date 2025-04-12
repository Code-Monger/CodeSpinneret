package tools

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestShell tests the shell tool with various operations
func TestShell(ctx context.Context, c client.MCPClient) error {
	// Create a session ID for testing
	sessionID := "shell-test-session-" + time.Now().Format("20060102-150405")

	// Initialize workspace first
	log.Printf("Initializing workspace for shell testing...")

	// Get current working directory as absolute path
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current working directory: %v", err)
		return err
	}

	workspaceReq := mcp.CallToolRequest{}
	workspaceReq.Params.Name = "workspace"
	workspaceReq.Params.Arguments = map[string]interface{}{
		"operation":  "initialize",
		"root_dir":   cwd,
		"user_task":  "Testing the shell tool",
		"session_id": sessionID,
	}

	workspaceResult, err := c.CallTool(ctx, workspaceReq)
	if err != nil {
		log.Printf("Failed to initialize workspace: %v", err)
		return err
	}

	if len(workspaceResult.Content) > 0 {
		if textContent, ok := workspaceResult.Content[0].(mcp.TextContent); ok {
			log.Printf("Workspace initialization result:\n%s", textContent.Text)
		}
	}

	// Define test cases for bash shell
	bashTestCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Initialize bash shell session",
			arguments: map[string]interface{}{
				"operation":  "initialize",
				"session_id": sessionID,
				"shell_type": "bash",
			},
		},
		{
			name: "Execute simple echo command",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "echo Hello from persistent bash shell!",
				"timeout":    5.0,
			},
		},
		{
			name: "Execute command that returns non-zero exit code",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "ls /nonexistent_directory && echo Success || echo 'Failed with exit code: $?'",
				"timeout":    5.0,
			},
		},
		{
			name: "Execute command that generates stderr",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "cat /nonexistent_file 2>&1",
				"timeout":    5.0,
			},
		},
		{
			name: "Set environment variable to test state persistence",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "export TEST_VAR='This is a persistent environment variable'",
				"timeout":    5.0,
			},
		},
		{
			name: "Echo environment variable to verify state persistence",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "echo $TEST_VAR",
				"timeout":    5.0,
			},
		},
		{
			name: "Execute multi-line command",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "for i in {1..3}; do\n  echo \"Line $i\"\ndone",
				"timeout":    5.0,
			},
		},
		{
			name: "Execute command with stdin input simulation",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "cat << EOF\nThis is line 1\nThis is line 2\nThis is line 3\nEOF",
				"timeout":    5.0,
			},
		},
		{
			name: "Check bash shell session status",
			arguments: map[string]interface{}{
				"operation":  "status",
				"session_id": sessionID,
			},
		},
		{
			name: "Close bash shell session",
			arguments: map[string]interface{}{
				"operation":  "close",
				"session_id": sessionID,
			},
		},
	}

	// Run bash test cases
	log.Printf("\n=== TESTING BASH SHELL ===\n")
	for _, tc := range bashTestCases {
		log.Printf("\nRunning bash shell test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "shell"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call shell: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Bash shell result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(1 * time.Second)
	}

	// Test error cases
	errorTestCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Try to use closed session",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": sessionID,
				"command":    "echo This should fail because session is closed",
				"timeout":    5.0,
			},
		},
		{
			name: "Missing required parameter",
			arguments: map[string]interface{}{
				"operation": "execute",
				// Missing session_id
				"command": "echo Missing session_id",
			},
		},
		{
			name: "Invalid operation",
			arguments: map[string]interface{}{
				"operation":  "invalid_operation",
				"session_id": sessionID,
			},
		},
	}

	// Run error test cases
	for _, tc := range errorTestCases {
		log.Printf("\nRunning shell error test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "shell"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Expected error received: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Unexpected success result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}

	// Test PowerShell
	log.Printf("\n=== TESTING POWERSHELL SHELL ===\n")
	powershellSessionID := "shell-test-powershell-" + time.Now().Format("20060102-150405")

	// Define test cases for PowerShell
	powershellTestCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Initialize PowerShell session",
			arguments: map[string]interface{}{
				"operation":  "initialize",
				"session_id": powershellSessionID,
				"shell_type": "powershell",
			},
		},
		{
			name: "Execute simple echo command in PowerShell",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": powershellSessionID,
				"command":    "Write-Host 'Hello from PowerShell!'",
				"timeout":    5.0,
			},
		},
		{
			name: "Execute command that returns non-zero exit code in PowerShell",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": powershellSessionID,
				"command":    "Get-Item NonExistentFile.txt -ErrorAction Stop; if (-not $?) { Write-Host \"Command failed with exit code: $LASTEXITCODE\" }",
				"timeout":    5.0,
			},
		},
		{
			name: "Set environment variable in PowerShell",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": powershellSessionID,
				"command":    "$env:PS_TEST_VAR = 'PowerShell environment variable'",
				"timeout":    5.0,
			},
		},
		{
			name: "Echo environment variable in PowerShell",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": powershellSessionID,
				"command":    "Write-Host $env:PS_TEST_VAR",
				"timeout":    5.0,
			},
		},
		{
			name: "Close PowerShell session",
			arguments: map[string]interface{}{
				"operation":  "close",
				"session_id": powershellSessionID,
			},
		},
	}

	// Run PowerShell test cases
	for _, tc := range powershellTestCases {
		log.Printf("\nRunning PowerShell test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "shell"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call shell for PowerShell: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("PowerShell result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(1 * time.Second)
	}

	// Test CMD shell
	log.Printf("\n=== TESTING CMD SHELL ===\n")
	cmdSessionID := "shell-test-cmd-" + time.Now().Format("20060102-150405")

	// Define test cases for CMD
	cmdTestCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Initialize CMD session",
			arguments: map[string]interface{}{
				"operation":  "initialize",
				"session_id": cmdSessionID,
				"shell_type": "cmd",
			},
		},
		{
			name: "Execute simple echo command in CMD",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": cmdSessionID,
				"command":    "echo Hello from CMD shell!",
				"timeout":    5.0,
			},
		},
		{
			name: "Execute command that returns non-zero exit code in CMD",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": cmdSessionID,
				"command":    "dir /nonexistent && echo Success || echo Failed with exit code %errorlevel%",
				"timeout":    5.0,
			},
		},
		{
			name: "Set environment variable in CMD",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": cmdSessionID,
				"command":    "set CMD_TEST_VAR=CMD environment variable",
				"timeout":    5.0,
			},
		},
		{
			name: "Echo environment variable in CMD",
			arguments: map[string]interface{}{
				"operation":  "execute",
				"session_id": cmdSessionID,
				"command":    "echo %CMD_TEST_VAR%",
				"timeout":    5.0,
			},
		},
		{
			name: "Close CMD session",
			arguments: map[string]interface{}{
				"operation":  "close",
				"session_id": cmdSessionID,
			},
		},
	}

	// Run CMD test cases
	for _, tc := range cmdTestCases {
		log.Printf("\nRunning CMD test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "shell"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call shell for CMD: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("CMD result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(1 * time.Second)
	}

	return nil
}
