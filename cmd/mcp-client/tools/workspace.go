package tools

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestWorkspace tests the workspace tool
func TestWorkspace(ctx context.Context, c client.MCPClient) error {
	// Test getting workspace info without initialization (should fail)
	log.Printf("Running workspace test: Get workspace info without initialization")
	uninitializedResult, err := testWorkspaceGetUninitialized(ctx, c)
	if err != nil {
		log.Printf("Workspace get without initialization failed as expected: %v", err)
	} else {
		log.Printf("Workspace get without initialization succeeded unexpectedly")
		if len(uninitializedResult.Content) > 0 {
			if textContent, ok := uninitializedResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Result: %s", textContent.Text)
			}
		}
	}

	// Test initializing a workspace
	log.Printf("Running workspace test: Initialize workspace")
	initResult, err := testWorkspaceInitialize(ctx, c)
	if err != nil {
		log.Printf("Workspace initialization failed: %v", err)
		return err
	}
	if len(initResult.Content) > 0 {
		if textContent, ok := initResult.Content[0].(mcp.TextContent); ok {
			log.Printf("Workspace result:\n%s", textContent.Text)
		}
	}

	// Test getting workspace info
	log.Printf("Running workspace test: Get workspace info")
	getResult, err := testWorkspaceGet(ctx, c)
	if err != nil {
		log.Printf("Workspace get failed: %v", err)
		return err
	}
	if len(getResult.Content) > 0 {
		if textContent, ok := getResult.Content[0].(mcp.TextContent); ok {
			log.Printf("Workspace result:\n%s", textContent.Text)
		}
	}

	// Test listing sessions
	log.Printf("Running workspace test: List sessions")
	listResult, err := testWorkspaceList(ctx, c)
	if err != nil {
		log.Printf("Workspace list failed: %v", err)
		return err
	}
	if len(listResult.Content) > 0 {
		if textContent, ok := listResult.Content[0].(mcp.TextContent); ok {
			log.Printf("Workspace result:\n%s", textContent.Text)
		}
	}

	// Test accessing workspace resource
	log.Printf("Reading workspace resource...")
	err = testWorkspaceResource(ctx, c)
	if err != nil {
		log.Printf("Failed to read workspace resource: %v", err)
		return err
	}

	return nil
}

// testWorkspaceInitialize tests initializing a workspace
func testWorkspaceInitialize(ctx context.Context, c client.MCPClient) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "workspace"
	req.Params.Arguments = map[string]interface{}{
		"operation":  "initialize",
		"root_dir":   ".",
		"user_task":  "Testing the workspace tool",
		"session_id": "test-session-1",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testWorkspaceGetUninitialized tests getting workspace info without initialization
func testWorkspaceGetUninitialized(ctx context.Context, c client.MCPClient) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "workspace"
	req.Params.Arguments = map[string]interface{}{
		"operation":  "get",
		"session_id": "nonexistent-session",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testWorkspaceGet tests getting workspace info
func testWorkspaceGet(ctx context.Context, c client.MCPClient) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "workspace"
	req.Params.Arguments = map[string]interface{}{
		"operation":  "get",
		"session_id": "test-session-1",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testWorkspaceList tests listing all sessions
func testWorkspaceList(ctx context.Context, c client.MCPClient) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "workspace"
	req.Params.Arguments = map[string]interface{}{
		"operation": "list",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testWorkspaceResource tests accessing the workspace resource
func testWorkspaceResource(ctx context.Context, c client.MCPClient) error {
	// Create the request
	req := mcp.ReadResourceRequest{}
	req.Params.URI = "workspace://info"

	// Call the resource
	result, err := c.ReadResource(ctx, req)
	if err != nil {
		return err
	}

	// Display the result
	if len(result.Contents) > 0 {
		if textContent, ok := result.Contents[0].(mcp.TextResourceContents); ok {
			log.Printf("Workspace Info:\n%s", textContent.Text)
		}
	}

	return nil
}
