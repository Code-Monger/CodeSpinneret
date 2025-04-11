package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFindCallers tests the find callers tool
func TestFindCallers(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_findcallers")

	// Create the test directory
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		log.Printf("Failed to create test directory: %v", err)
		return err
	}

	defer func() {
		// Clean up the test directory
		os.RemoveAll(testDir)
		log.Println("Test directory removed")
	}()

	log.Printf("Created test directory at: %s", testDir)

	// Create test files with function calls
	testFiles := map[string]string{
		"main.go": `package main

import (
	"fmt"
)

func main() {
	// Call the target function
	result := calculateTotal(10, 20)
	fmt.Println("Result:", result)
}

func calculateTotal(a, b int) int {
	return a + b
}
`,
		"utils.go": `package main

// Helper function that calls calculateTotal
func processNumbers(numbers []int) int {
	sum := 0
	for _, num := range numbers {
		sum += num
	}
	return calculateTotal(sum, 0)
}
`,
		"test.js": `// JavaScript test file
function testFunction() {
    // Call the target function
    const result = calculateTotal(5, 10);
    console.log("Result:", result);
}

function calculateTotal(a, b) {
    return a + b;
}
`,
	}

	// Write the test files
	for filename, content := range testFiles {
		filePath := filepath.Join(testDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			log.Printf("Failed to create test file %s: %v", filename, err)
			return err
		}
		log.Printf("Created test file: %s", filePath)
	}

	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Find callers of calculateTotal in Go files",
			arguments: map[string]interface{}{
				"function_name":      "calculateTotal",
				"search_directory":   testDir,
				"language":           "Go",
				"use_relative_paths": false,
				"recursive":          true,
			},
		},
		{
			name: "Find callers of calculateTotal in all languages",
			arguments: map[string]interface{}{
				"function_name":      "calculateTotal",
				"search_directory":   testDir,
				"use_relative_paths": false,
				"recursive":          true,
			},
		},
		{
			name: "Find callers with relative paths",
			arguments: map[string]interface{}{
				"function_name":      "calculateTotal",
				"search_directory":   testDir,
				"use_relative_paths": true,
				"recursive":          true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running find callers test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "findcallers"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call findcallers: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Find callers result:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
