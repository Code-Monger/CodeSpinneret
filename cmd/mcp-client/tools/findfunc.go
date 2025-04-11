package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFindFunc tests the findfunc tool
func TestFindFunc(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_findfunc")

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

	// Create test files with function definitions
	testFiles := map[string]string{
		"test_go.go": `package main

import (
	"fmt"
)

// calculateSum calculates the sum of two numbers
func calculateSum(a, b int) int {
	return a + b
}

func main() {
	result := calculateSum(10, 20)
	fmt.Println("Result:", result)
}
`,
		"test_js.js": `// JavaScript test file
function calculateSum(a, b) {
    return a + b;
}

// Test the function
const result = calculateSum(5, 10);
console.log("Result:", result);
`,
		"test_py.py": `# Python test file
def calculateSum(a, b):
    """Calculate the sum of two numbers."""
    return a + b

# Test the function
result = calculateSum(5, 10)
print("Result:", result)
`,
		"test_go_package.go": `package calculator

// calculateSum calculates the sum of two numbers
func calculateSum(a, b int) int {
	return a + b
}

// multiplyNumbers multiplies two numbers
func multiplyNumbers(a, b int) int {
	return a * b
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
			name: "Find function in all languages",
			arguments: map[string]interface{}{
				"function_name":      "calculateSum",
				"search_directory":   testDir,
				"use_relative_paths": true,
			},
		},
		{
			name: "Find function in Go only",
			arguments: map[string]interface{}{
				"function_name":      "calculateSum",
				"search_directory":   testDir,
				"language":           "Go",
				"use_relative_paths": true,
			},
		},
		{
			name: "Find function with package filter",
			arguments: map[string]interface{}{
				"function_name":      "calculateSum",
				"package_name":       "calculator",
				"search_directory":   testDir,
				"use_relative_paths": true,
			},
		},
		{
			name: "Find function that doesn't exist",
			arguments: map[string]interface{}{
				"function_name":      "nonExistentFunction",
				"search_directory":   testDir,
				"use_relative_paths": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running findfunc test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "findfunc"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call findfunc: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Find function result:\n%s", textContent.Text)
			}
		}
	}

	// Test using findfunc with funcdef
	log.Printf("Testing findfunc with funcdef integration")

	// First, find the function
	findReq := mcp.CallToolRequest{}
	findReq.Params.Name = "findfunc"
	findReq.Params.Arguments = map[string]interface{}{
		"function_name":      "calculateSum",
		"search_directory":   testDir,
		"language":           "Go",
		"use_relative_paths": false,
	}

	findResult, err := c.CallTool(ctx, findReq)
	if err != nil {
		log.Printf("Failed to call findfunc: %v", err)
		return nil
	}

	if len(findResult.Content) > 0 {
		if textContent, ok := findResult.Content[0].(mcp.TextContent); ok {
			log.Printf("Find function result:\n%s", textContent.Text)
		}

		// Now use funcdef to get the function definition
		// In a real scenario, we would parse the JSON result from findfunc
		// For this test, we'll just use the first file directly
		defReq := mcp.CallToolRequest{}
		defReq.Params.Name = "funcdef"
		defReq.Params.Arguments = map[string]interface{}{
			"operation":     "get",
			"function_name": "calculateSum",
			"file_path":     filepath.Join(testDir, "test_go.go"),
			"language":      "Go",
		}

		defResult, err := c.CallTool(ctx, defReq)
		if err != nil {
			log.Printf("Failed to call funcdef: %v", err)
			return nil
		}

		if len(defResult.Content) > 0 {
			if textContent, ok := defResult.Content[0].(mcp.TextContent); ok {
				log.Printf("Function definition result:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
