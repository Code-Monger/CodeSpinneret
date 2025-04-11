package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFuncDefWithComplexStrings tests the function definition tool with complex string literals
func TestFuncDefWithComplexStrings(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_funcdef_complex")

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

	// Create test file with complex string literals
	testFile := filepath.Join(testDir, "test_complex_strings.go")
	fileContent := `package main

import (
	"fmt"
)

func processComplexStrings() {
	// String with braces and comment-like content
	str1 := "This string has braces { } and comment-like content // { }"
	
	// Raw string with braces and comment-like content
	str2 := ` + "`" + `This raw string has braces { } 
	and comment-like content // { }
	/* 
	   This looks like a multi-line comment
	   but it's actually inside a raw string
	   {
	      nested {
	         braces
	      }
	   }
	*/` + "`" + `
	
	// String with escaped quotes and braces
	str3 := "String with \"escaped quotes\" and braces { } and a comment-like part // not a comment"
	
	// String with nested quotes
	str4 := "String with 'single quotes' and braces { }"
	
	// String with escaped braces
	str5 := "String with escaped braces \\{ \\}"
	
	fmt.Println(str1, str2, str3, str4, str5)
}

func main() {
	processComplexStrings()
}
`

	err = os.WriteFile(testFile, []byte(fileContent), 0644)
	if err != nil {
		log.Printf("Failed to create test file: %v", err)
		return err
	}
	log.Printf("Created test file: %s", testFile)

	// Define replacement content
	replacementContent := `func processComplexStrings() {
	// Modified function with complex string literals
	
	// Added a new variable
	newVar := 42
	
	// String with braces and comment-like content
	str1 := "This string has braces { } and comment-like content // { }"
	
	// Raw string with braces and comment-like content
	str2 := ` + "`" + `This raw string has braces { } 
	and comment-like content // { }
	/* 
	   This looks like a multi-line comment
	   but it's actually inside a raw string
	   {
	      nested {
	         braces
	      }
	   }
	*/` + "`" + `
	
	// String with escaped quotes and braces
	str3 := "String with \"escaped quotes\" and braces { } and a comment-like part // not a comment"
	
	// String with nested quotes
	str4 := "String with 'single quotes' and braces { }"
	
	// String with escaped braces
	str5 := "String with escaped braces \\{ \\}"
	
	fmt.Println(str1, str2, str3, str4, str5, newVar)
}`

	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Get function with complex string literals",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "processComplexStrings",
				"file_path":     testFile,
				"language":      "Go",
			},
		},
		{
			name: "Replace function with complex string literals",
			arguments: map[string]interface{}{
				"operation":           "replace",
				"function_name":       "processComplexStrings",
				"file_path":           testFile,
				"language":            "Go",
				"replacement_content": replacementContent,
			},
		},
		{
			name: "Get replaced function with complex string literals",
			arguments: map[string]interface{}{
				"operation":     "get",
				"function_name": "processComplexStrings",
				"file_path":     testFile,
				"language":      "Go",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running function definition test with complex string literals: %s", tc.name)

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
				log.Printf("Function definition result with complex string literals:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
