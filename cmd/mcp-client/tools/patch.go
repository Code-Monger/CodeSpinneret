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

// TestPatch tests the patch tool with various operations
func TestPatch(ctx context.Context, c client.MCPClient) error {
	// Create a temporary test file
	tempDir := os.TempDir()
	testFilePath := filepath.Join(tempDir, "mcp_test_patch.txt")

	// Create original content
	originalContent := `This is a test file for patching.
It contains multiple lines of text.
This line will be modified.
This line will be removed.
This is the last line of the file.`

	err := ioutil.WriteFile(testFilePath, []byte(originalContent), 0644)
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
	log.Printf("Original content:\n%s", originalContent)

	// Create a patch in unified diff format
	patchContent := `--- mcp_test_patch.txt
+++ mcp_test_patch.txt
@@ -1,5 +1,6 @@
	This is a test file for patching.
	It contains multiple lines of text.
-This line will be modified.
-This line will be removed.
+This line has been modified.
+This is a new line that was added.
+Another new line was added here.
	This is the last line of the file.`

	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Dry run patch",
			arguments: map[string]interface{}{
				"patch_content":    patchContent,
				"target_directory": tempDir,
				"strip_level":      0.0,
				"dry_run":          true,
			},
		},
		{
			name: "Apply patch",
			arguments: map[string]interface{}{
				"patch_content":    patchContent,
				"target_directory": tempDir,
				"strip_level":      0.0,
				"dry_run":          false,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running patch test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "patch"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call patch: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Patch result:\n%s", textContent.Text)
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

	log.Printf("Final file content after patching:\n%s", string(modifiedContent))

	return nil
}
