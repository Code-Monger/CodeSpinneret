// Package tools provides test functions for MCP tools
package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFuncDefWithCommentsInternal tests the function definition tool with tricky comments
func TestFuncDefWithCommentsInternal(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_funcdef_comments")

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

	// Create test files with function definitions and tricky comments
	testFiles := map[string]string{
		"test_go_comments.go": `package main

import (
	"fmt"
)

// calculateSum calculates the sum of two numbers
// { This comment has a brace that could confuse the parser
func calculateSum(a, b int) int {
	// Another comment with a brace }
	/*
	   Multi-line comment with braces
	   {
	   }
	*/
	return a + b // Inline comment with }
}

/*
} This multi-line comment starts with a closing brace
*/

func main() {
	result := calculateSum(10, 20)
	fmt.Println("Result:", result)
}
`,
		"test_cpp_comments.cpp": `#include <iostream>

// Function prototype
int calculateSum(int a, int b); // { Comment with brace

int main() {
    // Comment with brace }
    int result = calculateSum(10, 20);
    std::cout << "Result: " << result << std::endl;
    return 0;
}

/*
   Multi-line comment with braces
   {
   }
*/
// Function implementation
int calculateSum(int a, int b) {
    /* } Tricky comment with closing brace at start */
    return a + b; // Inline comment with }
}
`,
		"test_js_comments.js": `// JavaScript test file with tricky comments

// Function with comments that have braces
// { This comment has an opening brace
function calculateSum(a, b) {
    // } This comment has a closing brace
    /*
       Multi-line comment with braces
       {
       }
    */
    return a + b; // Inline comment with }
}

/*
} This multi-line comment starts with a closing brace
*/

// Test the function
const result = calculateSum(5, 10);
console.log("Result:", result);
`,
		"test_py_comments.py": `# Python test file with tricky comments

# Function with comments that have indentation and braces
# { This comment has an opening brace
def calculateSum(a, b):
    """
    Calculate the sum of two numbers.
    
    This docstring has braces:
    {
    }
    """
    # } This comment has a closing brace
    return a + b  # Inline comment with }

# This comment is at the same indentation level as the function
# but shouldn't be considered part of it

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
			name: "Get Go function with tricky comments",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test_go_comments.go"),
				"language":      "Go",
			},
		},
		{
			name: "Get C++ function with tricky comments",
			arguments: map[string]interface{}{
				"operation":         "get",
				"function_name":     "calculateSum",
				"file_path":         filepath.Join(testDir, "test_cpp_comments.cpp"),
				"language":          "C/C++",
				"include_prototype": true,
			},
		},
		{
			name: "Get JavaScript function with tricky comments",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test_js_comments.js"),
				"language":      "JavaScript",
			},
		},
		{
			name: "Get Python function with tricky comments",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test_py_comments.py"),
				"language":      "Python",
			},
		},
		{
			name: "Replace Go function with tricky comments",
			arguments: map[string]interface{}{
				"operation":     "replace",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test_go_comments.go"),
				"language":      "Go",
				"replacement_content": `// calculateSum calculates the sum of two numbers and adds a bonus
// { This comment has a brace that could confuse the parser
func calculateSum(a, b int) int {
	// Another comment with a brace }
	/*
	   Multi-line comment with braces
	   {
	   }
	*/
	bonus := 5 // New variable
	return a + b + bonus // Modified return with }
}`,
			},
		},
		{
			name: "Get replaced Go function with tricky comments",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "calculateSum",
				"file_path":     filepath.Join(testDir, "test_go_comments.go"),
				"language":      "Go",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running function definition test with comments: %s", tc.name)

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
				log.Printf("Function definition result with comments:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
