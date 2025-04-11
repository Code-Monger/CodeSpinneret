package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestLineCount tests the line count tool
func TestLineCount(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test file
	tempDir := os.TempDir()
	testFilePath := filepath.Join(tempDir, "mcp_test_linecount.txt")

	// Create test content with known line, word, and character counts
	testContent := "This is line one.\nThis is line two.\nThis is line three.\nThis is line four.\nThis is line five."
	// 5 lines, 20 words, 95 characters (including newlines)

	// Write the test content to the file
	err := os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		log.Printf("Failed to create test file: %v", err)
		return err
	}

	defer func() {
		// Clean up the test file
		os.Remove(testFilePath)
		log.Println("Test file removed")
	}()

	log.Printf("Created test file at: %s", testFilePath)
	log.Printf("Test content:\n%s", testContent)

	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Count lines only",
			arguments: map[string]interface{}{
				"file_path":   testFilePath,
				"count_lines": true,
				"count_words": false,
				"count_chars": false,
			},
		},
		{
			name: "Count words only",
			arguments: map[string]interface{}{
				"file_path":   testFilePath,
				"count_lines": false,
				"count_words": true,
				"count_chars": false,
			},
		},
		{
			name: "Count characters only",
			arguments: map[string]interface{}{
				"file_path":   testFilePath,
				"count_lines": false,
				"count_words": false,
				"count_chars": true,
			},
		},
		{
			name: "Count all (lines, words, characters)",
			arguments: map[string]interface{}{
				"file_path":   testFilePath,
				"count_lines": true,
				"count_words": true,
				"count_chars": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running line count test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "linecount"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call linecount: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Line count result:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
