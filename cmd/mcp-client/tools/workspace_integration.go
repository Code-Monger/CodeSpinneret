package tools

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestWorkspaceIntegration tests the integration between workspace and other tools
func TestWorkspaceIntegration(ctx context.Context, c client.MCPClient) error {
	// Create a test file for the tests
	testFilePath := "test_workspace_integration.txt"
	err := os.WriteFile(testFilePath, []byte("Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"), 0644)
	if err != nil {
		log.Printf("Error creating test file: %v", err)
		return err
	}
	defer os.Remove(testFilePath)

	// Test using linecount without initializing workspace
	log.Printf("Running integration test: Using linecount without workspace initialization")
	linecountResult, err := testLinecountWithoutWorkspace(ctx, c, testFilePath)
	if err != nil {
		log.Printf("Linecount without workspace failed: %v", err)
	} else {
		log.Printf("Linecount without workspace succeeded")
		if len(linecountResult.Content) > 0 {
			if textContent, ok := linecountResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Result: %s", textContent.Text)
			}
		}
	}

	// Test using patch without initializing workspace
	log.Printf("Running integration test: Using patch without workspace initialization")
	patchResult, err := testPatchWithoutWorkspace(ctx, c, testFilePath)
	if err != nil {
		log.Printf("Patch without workspace failed: %v", err)
	} else {
		log.Printf("Patch without workspace succeeded")
		if len(patchResult.Content) > 0 {
			if textContent, ok := patchResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Result: %s", textContent.Text)
			}
		}
	}

	// Initialize workspace
	log.Printf("Running integration test: Initializing workspace")
	workspaceResult, err := testWorkspaceInitializeForIntegration(ctx, c)
	if err != nil {
		log.Printf("Workspace initialization failed: %v", err)
		return err
	}
	log.Printf("Workspace initialized successfully")
	if len(workspaceResult.Content) > 0 {
		if textContent, ok := workspaceResult.Content[0].(mcp.TextContent); ok {
			log.Printf("Result: %s", textContent.Text)
		}
	}

	// Test using linecount with initialized workspace
	log.Printf("Running integration test: Using linecount with workspace initialization")
	linecountWithWorkspaceResult, err := testLinecountWithWorkspace(ctx, c, testFilePath)
	if err != nil {
		log.Printf("Linecount with workspace failed: %v", err)
	} else {
		log.Printf("Linecount with workspace succeeded")
		if len(linecountWithWorkspaceResult.Content) > 0 {
			if textContent, ok := linecountWithWorkspaceResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Result: %s", textContent.Text)
			}
		}
	}

	// Test using patch with initialized workspace
	log.Printf("Running integration test: Using patch with workspace initialization")
	patchWithWorkspaceResult, err := testPatchWithWorkspace(ctx, c, testFilePath)
	if err != nil {
		log.Printf("Patch with workspace failed: %v", err)
	} else {
		log.Printf("Patch with workspace succeeded")
		if len(patchWithWorkspaceResult.Content) > 0 {
			if textContent, ok := patchWithWorkspaceResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Result: %s", textContent.Text)
			}
		}
	}

	// Test using RAG without initializing workspace
	log.Printf("Running integration test: Using RAG without workspace initialization")
	ragResult, err := testRAGWithoutWorkspace(ctx, c)
	if err != nil {
		log.Printf("RAG without workspace failed: %v", err)
	} else {
		log.Printf("RAG without workspace succeeded")
		if len(ragResult.Content) > 0 {
			if textContent, ok := ragResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Result: %s", textContent.Text)
			}
		}
	}

	return nil
}

// testWorkspaceInitializeForIntegration tests initializing the workspace
func testWorkspaceInitializeForIntegration(ctx context.Context, c client.MCPClient) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "workspace"
	req.Params.Arguments = map[string]interface{}{
		"operation":  "initialize",
		"root_dir":   ".",
		"user_task":  "Testing workspace integration",
		"session_id": "test-session-1",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testLinecountWithoutWorkspace tests the linecount tool without initializing workspace
func testLinecountWithoutWorkspace(ctx context.Context, c client.MCPClient, filePath string) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "linecount"
	req.Params.Arguments = map[string]interface{}{
		"file_path":   filePath,
		"count_lines": true,
		"count_words": true,
		"count_chars": true,
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testLinecountWithWorkspace tests the linecount tool with initialized workspace
func testLinecountWithWorkspace(ctx context.Context, c client.MCPClient, filePath string) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "linecount"
	req.Params.Arguments = map[string]interface{}{
		"file_path":   filePath,
		"count_lines": true,
		"count_words": true,
		"count_chars": true,
		"session_id":  "test-session-1",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testPatchWithoutWorkspace tests the patch tool without initializing workspace
func testPatchWithoutWorkspace(ctx context.Context, c client.MCPClient, filePath string) (*mcp.CallToolResult, error) {
	// Create a patch content
	patchContent := fmt.Sprintf(`--- %s	2025-04-11 18:35:35.000000000 -0500
+++ %s	2025-04-11 18:35:35.000000000 -0500
@@ -1,5 +1,5 @@
 Line 1
 Line 2
-Line 3
+Line 3 (modified)
 Line 4
 Line 5`, filePath, filePath)

	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "patch"
	req.Params.Arguments = map[string]interface{}{
		"patch_content":    patchContent,
		"target_directory": ".",
		"dry_run":          true,
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testPatchWithWorkspace tests the patch tool with initialized workspace
func testPatchWithWorkspace(ctx context.Context, c client.MCPClient, filePath string) (*mcp.CallToolResult, error) {
	// Create a patch content
	patchContent := fmt.Sprintf(`--- %s	2025-04-11 18:35:35.000000000 -0500
+++ %s	2025-04-11 18:35:35.000000000 -0500
@@ -1,5 +1,5 @@
 Line 1
 Line 2
-Line 3
+Line 3 (modified with workspace)
 Line 4
 Line 5`, filePath, filePath)

	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "patch"
	req.Params.Arguments = map[string]interface{}{
		"patch_content":    patchContent,
		"target_directory": ".",
		"dry_run":          true,
		"session_id":       "test-session-1",
	}

	// Call the tool
	return c.CallTool(ctx, req)
}

// testRAGWithoutWorkspace tests the RAG tool without initializing workspace
func testRAGWithoutWorkspace(ctx context.Context, c client.MCPClient) (*mcp.CallToolResult, error) {
	// Create the request
	req := mcp.CallToolRequest{}
	req.Params.Name = "rag"
	req.Params.Arguments = map[string]interface{}{
		"operation":     "index",
		"repo_path":     ".",
		"file_patterns": []interface{}{"*.go", "*.md"},
	}

	// Call the tool
	return c.CallTool(ctx, req)
}
