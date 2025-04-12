package tools

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestRAG tests the RAG tool with various operations
func TestRAG(ctx context.Context, c client.MCPClient) error {
	// Initialize workspace first
	log.Printf("Initializing workspace for RAG testing...")
	sessionID := "rag-test-session"

	// Get current working directory as absolute path
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current working directory: %v", err)
		return err
	}

	// Initialize workspace
	workspaceReq := mcp.CallToolRequest{}
	workspaceReq.Params.Name = "workspace"
	workspaceReq.Params.Arguments = map[string]interface{}{
		"operation":  "initialize",
		"root_dir":   cwd,
		"user_task":  "Testing the enhanced RAG tool",
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
				"session_id":    sessionID,
				"file_patterns": []interface{}{"*.go", "*.md"},
			},
		},
		{
			name: "Query repository - General query",
			arguments: map[string]interface{}{
				"operation":   "query",
				"repo_path":   ".",
				"session_id":  sessionID,
				"query":       "how to implement a service",
				"num_results": 3.0,
			},
		},
		{
			name: "Query repository - Function name query",
			arguments: map[string]interface{}{
				"operation":   "query",
				"repo_path":   ".",
				"session_id":  sessionID,
				"query":       "Show me the findHunkLocation function in patch.go",
				"num_results": 3.0,
			},
		},
		{
			name: "Query repository - Exact function signature query",
			arguments: map[string]interface{}{
				"operation":   "query",
				"repo_path":   ".",
				"session_id":  sessionID,
				"query":       "func findHunkLocation(lines []string, hunk Hunk) int",
				"num_results": 3.0,
			},
		},
		{
			name: "Query repository - Function extraction test",
			arguments: map[string]interface{}{
				"operation":   "query",
				"repo_path":   ".",
				"session_id":  sessionID,
				"query":       "extractSnippet function in rag package",
				"num_results": 3.0,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("\nRunning RAG test: %s", tc.name)

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
