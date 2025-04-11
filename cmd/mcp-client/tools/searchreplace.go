package tools

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestSearchReplace tests the search replace tool with various operations
func TestSearchReplace(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test file
	tempDir := os.TempDir()
	testFilePath := filepath.Join(tempDir, "mcp_test_search_replace.txt")

	testContent := `This is a test file for search and replace.
It contains multiple lines of text.
We will search for specific patterns and replace them.
This line has the word 'test' in it twice for testing.
The end of the test file.`

	err := ioutil.WriteFile(testFilePath, []byte(testContent), 0644)
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

	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Simple string replacement (preview)",
			arguments: map[string]interface{}{
				"directory":      filepath.Dir(testFilePath),
				"file_pattern":   filepath.Base(testFilePath),
				"search_pattern": "test",
				"replacement":    "EXAMPLE",
				"use_regex":      false,
				"recursive":      false,
				"preview":        true,
				"case_sensitive": true,
			},
		},
		{
			name: "Regex replacement (preview)",
			arguments: map[string]interface{}{
				"directory":      filepath.Dir(testFilePath),
				"file_pattern":   filepath.Base(testFilePath),
				"search_pattern": "t[a-z]{3}",
				"replacement":    "MATCH",
				"use_regex":      true,
				"recursive":      false,
				"preview":        true,
				"case_sensitive": false,
			},
		},
		{
			name: "Actual replacement",
			arguments: map[string]interface{}{
				"directory":      filepath.Dir(testFilePath),
				"file_pattern":   filepath.Base(testFilePath),
				"search_pattern": "line",
				"replacement":    "ROW",
				"use_regex":      false,
				"recursive":      false,
				"preview":        false,
				"case_sensitive": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running search replace test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "searchreplace"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call searchreplace: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Search replace result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}

	// Read the modified file to show changes
	modifiedContent, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		log.Printf("Failed to read modified file: %v", err)
		return err
	}

	log.Printf("Final file content after modifications:\n%s", string(modifiedContent))

	return nil
}
