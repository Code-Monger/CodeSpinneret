package tools

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestSpellCheck tests the spellcheck tool
func TestSpellCheck(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test directory
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "mcp_test_spellcheck")

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

	// Create test files with spelling issues
	testFiles := map[string]string{
		"test_comments.go": `package main

import (
	"fmt"
)

// This is a coment with a speling mistake
func main() {
	// Another coment with a mispelled word
	fmt.Println("Hello, World!")
}
`,
		"test_strings.go": `package main

import (
	"fmt"
)

func main() {
	// String with spelling mistakes
	message := "This is a mesage with a speling mistake"
	fmt.Println(message)
}
`,
		"test_identifiers.go": `package main

import (
	"fmt"
)

func main() {
	// Variable with spelling mistake
	userAcount := "John"
	fmt.Println(userAcount)

	// Function with spelling mistake
	displayMessge("Hello")
}

func displayMessge(text string) {
	fmt.Println(text)
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
			name: "Check all types",
			arguments: map[string]interface{}{
				"path":               testDir,
				"check_comments":     true,
				"check_strings":      true,
				"check_identifiers":  true,
				"recursive":          true,
				"use_relative_paths": true,
			},
		},
		{
			name: "Check comments only",
			arguments: map[string]interface{}{
				"path":               testDir,
				"check_comments":     true,
				"check_strings":      false,
				"check_identifiers":  false,
				"recursive":          true,
				"use_relative_paths": true,
			},
		},
		{
			name: "Check with custom dictionary",
			arguments: map[string]interface{}{
				"path":               testDir,
				"check_comments":     true,
				"check_strings":      true,
				"check_identifiers":  true,
				"recursive":          true,
				"use_relative_paths": true,
				"custom_dictionary":  []interface{}{"speling", "coment"},
			},
		},
		{
			name: "Check specific file",
			arguments: map[string]interface{}{
				"path":               filepath.Join(testDir, "test_identifiers.go"),
				"check_comments":     true,
				"check_strings":      true,
				"check_identifiers":  true,
				"use_relative_paths": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running spellcheck test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "spellcheck"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call spellcheck: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Spellcheck result:\n%s", textContent.Text)
			}
		}
	}

	return nil
}
