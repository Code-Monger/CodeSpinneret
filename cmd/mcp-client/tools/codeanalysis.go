package tools

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestCodeAnalysis tests the code analysis tool with various operations
func TestCodeAnalysis(ctx context.Context, c client.MCPClient) error {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Analyze file",
			arguments: map[string]interface{}{
				"operation": "analyze_file",
				"file_path": "cmd/mcp-server/main.go",
			},
		},
		{
			name: "Analyze directory",
			arguments: map[string]interface{}{
				"operation":      "analyze_directory",
				"directory_path": "pkg/calculator",
				"file_patterns":  []interface{}{"*.go"},
				"recursive":      true,
			},
		},
		{
			name: "Find issues",
			arguments: map[string]interface{}{
				"operation":   "find_issues",
				"target_path": "pkg/test/test.go",
				"issue_types": []interface{}{"complexity", "comments"},
				"severity":    "medium",
			},
		},
		{
			name: "Suggest improvements",
			arguments: map[string]interface{}{
				"operation":         "suggest_improvements",
				"target_path":       "pkg/serverinfo/serverinfo.go",
				"improvement_types": []interface{}{"readability", "maintainability"},
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running code analysis test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "codeanalysis"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call codeanalysis: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Code analysis result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(1 * time.Second)
	}

	return nil
}
