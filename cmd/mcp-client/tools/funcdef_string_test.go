package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFuncDefWithStrings tests the function definition tool with string literals
func TestFuncDefWithStrings(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_funcdef_strings")

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

	// Create test files with function definitions and string literals
	testFiles := map[string]string{
		"test_go_strings.go": `package main

import (
	"fmt"
)

func processStrings() {
	// Double-quoted string with braces
	str1 := "This string has braces: { } and more {{}}"
	
	// Raw string with braces (backtick-quoted)
	str2 := ` + "`" + `This raw string has braces: { } 
	and more {{}} on multiple
	lines with indentation {
		nested {
			content
		}
	}` + "`" + `
	
	// Single-quoted rune with brace
	char1 := '{'
	char2 := '}'
	
	// String with escaped quotes
	str3 := "String with \"escaped quotes\" and braces { }"
	
	// String with comment-like content
	str4 := "This looks like a comment: // but it's not"
	str5 := "This looks like a comment: /* but it's not */"
	
	fmt.Println(str1, str2, char1, char2, str3, str4, str5)
}

func main() {
	processStrings()
}
`,
		"test_cpp_strings.cpp": `#include <iostream>
#include <string>

void processStrings() {
    // Double-quoted string with braces
    std::string str1 = "This string has braces: { } and more {{}}";
    
    // Character literals with braces
    char char1 = '{';
    char char2 = '}';
    
    // String with escaped quotes
    std::string str2 = "String with \"escaped quotes\" and braces { }";
    
    // String with comment-like content
    std::string str3 = "This looks like a comment: // but it's not";
    std::string str4 = "This looks like a comment: /* but it's not */";
    
    // Multi-line string using backslash continuation
    std::string str5 = "This is a multi-line string \\
    with braces { } \\
    and more {{}} \\
    on multiple lines";
    
    // Raw string literals (C++11)
    std::string str6 = R"(This is a raw string with braces: { }
    and more {{}} on multiple
    lines with indentation {
        nested {
            content
        }
    })";
    
    std::cout << str1 << str2 << char1 << char2 << str3 << str4 << str5 << str6 << std::endl;
}

int main() {
    processStrings();
    return 0;
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
			name: "Get Go function with string literals",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "processStrings",
				"file_path":     filepath.Join(testDir, "test_go_strings.go"),
				"language":      "Go",
			},
		},
		{
			name: "Get C++ function with string literals",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "processStrings",
				"file_path":     filepath.Join(testDir, "test_cpp_strings.cpp"),
				"language":      "C/C++",
			},
		},
		{
			name: "Replace Go function with string literals",
			arguments: map[string]interface{}{
				"operation":     "replace",
				"function_name": "processStrings",
				"file_path":     filepath.Join(testDir, "test_go_strings.go"),
				"language":      "Go",
				"replacement_content": `func processStrings() {
	// Modified function with string literals
	str1 := "This string has braces: { } and more {{}}"
	
	// Added a new variable
	newVar := 42
	
	// Raw string with braces (backtick-quoted)
	str2 := ` + "`" + `This raw string has braces: { } 
	and more {{}} on multiple
	lines with indentation {
		nested {
			content
		}
	}` + "`" + `
	
	// Single-quoted rune with brace
	char1 := '{'
	char2 := '}'
	
	// String with escaped quotes
	str3 := "String with \"escaped quotes\" and braces { }"
	
	// String with comment-like content
	str4 := "This looks like a comment: // but it's not"
	str5 := "This looks like a comment: /* but it's not */"
	
	fmt.Println(str1, str2, char1, char2, str3, str4, str5, newVar)
}`,
			},
		},
		{
			name: "Get replaced Go function with string literals",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "processStrings",
				"file_path":     filepath.Join(testDir, "test_go_strings.go"),
				"language":      "Go",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running function definition test with string literals: %s", tc.name)

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
				log.Printf("Function definition result with string literals:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
