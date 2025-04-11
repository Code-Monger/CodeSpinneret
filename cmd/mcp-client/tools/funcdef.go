package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFuncDef tests the function definition tool
func TestFuncDef(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_funcdef")

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
		"test.go": `package main

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
		"test.cpp": `#include <iostream>

// Function prototype
int calculateSum(int a, int b);

int main() {
    int result = calculateSum(10, 20);
    std::cout << "Result: " << result << std::endl;
    return 0;
}

// Function implementation
int calculateSum(int a, int b) {
    return a + b;
}
`,
		"test.js": `// JavaScript test file
function calculateSum(a, b) {
    return a + b;
}

// Test the function
const result = calculateSum(5, 10);
console.log("Result:", result);
`,
		"test.py": `# Python test file
def calculateSum(a, b):
    """Calculate the sum of two numbers."""
    return a + b

# Test the function
result = calculateSum(5, 10)
print("Result:", result)
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
			name: "Get Go function",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test.go"),
				"language":      "Go",
			},
		},
		{
			name: "Get C++ function with prototype",
			arguments: map[string]interface{}{
				"operation":         "get",
				"function_name":     "calculateSum",
				"file_path":         filepath.Join(testDir, "test.cpp"),
				"language":          "C/C++",
				"include_prototype": true,
			},
		},
		{
			name: "Get JavaScript function",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test.js"),
			},
		},
		{
			name: "Get Python function",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test.py"),
			},
		},
		{
			name: "Replace Go function",
			arguments: map[string]interface{}{
				"operation":     "replace",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test.go"),
				"language":      "Go",
				"replacement_content": `// calculateSum calculates the sum of two numbers and adds a bonus
func calculateSum(a, b int) int {
	bonus := 5
	return a + b + bonus
}`,
			},
		},
		{
			name: "Get replaced Go function",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test.go"),
				"language":      "Go",
			},
		},
		{
			name: "Replace C++ function implementation",
			arguments: map[string]interface{}{
				"operation":     "replace",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test.cpp"),
				"language":      "C/C++",
				"replacement_content": `// Function implementation with bonus
int calculateSum(int a, int b) {
    int bonus = 5;
    return a + b + bonus;
}`,
			},
		},
		{
			name: "Get replaced C++ function",
			arguments: map[string]interface{}{
				"operation":         "get",
				"function_name":     "calculateSum",
				"file_path":         filepath.Join(testDir, "test.cpp"),
				"language":          "C/C++",
				"include_prototype": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running function definition test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "funcdef"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call funcdef: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Function definition result:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
