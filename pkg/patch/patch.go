package patch

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandlePatch is the handler function for the patch tool
func HandlePatch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract patch content
	patchContent, ok := arguments["patch_content"].(string)
	if !ok {
		return nil, fmt.Errorf("patch_content must be a string")
	}

	// Extract target directory
	targetDir, _ := arguments["target_directory"].(string)
	if targetDir == "" {
		targetDir = "." // Default to current directory
	}

	// Extract strip level
	stripLevel := 0 // Default strip level
	if stripLevelFloat, ok := arguments["strip_level"].(float64); ok {
		stripLevel = int(stripLevelFloat)
	}

	// Extract dry run flag
	dryRun := false // Default to actually applying the patch
	if dryRunBool, ok := arguments["dry_run"].(bool); ok {
		dryRun = dryRunBool
	}

	// Apply the patch
	result, err := applyPatch(patchContent, targetDir, stripLevel, dryRun)
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Patch Application Results:\n\n")
	resultText += fmt.Sprintf("Target directory: %s\n", targetDir)
	resultText += fmt.Sprintf("Strip level: %d\n", stripLevel)
	resultText += fmt.Sprintf("Dry run: %t\n\n", dryRun)

	if len(result.FilesPatched) > 0 {
		resultText += "Files patched:\n"
		for _, file := range result.FilesPatched {
			resultText += fmt.Sprintf("- %s\n", file)
		}
		resultText += "\n"
	}

	if len(result.FilesSkipped) > 0 {
		resultText += "Files skipped:\n"
		for file, reason := range result.FilesSkipped {
			resultText += fmt.Sprintf("- %s: %s\n", file, reason)
		}
		resultText += "\n"
	}

	resultText += fmt.Sprintf("Hunks applied: %d\n", result.HunksApplied)
	resultText += fmt.Sprintf("Hunks failed: %d\n", result.HunksFailed)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// PatchResult represents the result of applying a patch
type PatchResult struct {
	FilesPatched []string
	FilesSkipped map[string]string
	HunksApplied int
	HunksFailed  int
}

// FilePatch represents a patch for a single file
type FilePatch struct {
	SourceFile string
	TargetFile string
	Hunks      []Hunk
}

// Hunk represents a single hunk in a patch
type Hunk struct {
	SourceStart int
	SourceLines int
	TargetStart int
	TargetLines int
	Context     []string
	Added       []string
	Removed     []string
}

// applyPatch applies a patch to files in the target directory
func applyPatch(patchContent, targetDir string, stripLevel int, dryRun bool) (*PatchResult, error) {
	result := &PatchResult{
		FilesPatched: []string{},
		FilesSkipped: make(map[string]string),
		HunksApplied: 0,
		HunksFailed:  0,
	}

	// Parse the patch content
	patches, err := parsePatch(patchContent)
	if err != nil {
		return nil, err
	}

	// Apply each file patch
	for _, filePatch := range patches {
		// Apply strip level to the file path
		targetPath := applyStripLevel(filePatch.TargetFile, stripLevel)

		// Resolve the full path
		fullPath := filepath.Join(targetDir, targetPath)

		// Check if the file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			result.FilesSkipped[targetPath] = "file does not exist"
			continue
		}

		// Read the file content
		content, err := ioutil.ReadFile(fullPath)
		if err != nil {
			result.FilesSkipped[targetPath] = fmt.Sprintf("error reading file: %v", err)
			continue
		}

		// Apply the hunks to the file content
		newContent, hunksApplied, hunksFailed := applyHunks(string(content), filePatch.Hunks)
		result.HunksApplied += hunksApplied
		result.HunksFailed += hunksFailed

		// If no hunks were applied, skip the file
		if hunksApplied == 0 {
			result.FilesSkipped[targetPath] = "no hunks applied"
			continue
		}

		// Write the new content to the file (unless dry run)
		if !dryRun {
			err = ioutil.WriteFile(fullPath, []byte(newContent), 0644)
			if err != nil {
				result.FilesSkipped[targetPath] = fmt.Sprintf("error writing file: %v", err)
				continue
			}
		}

		// Add the file to the list of patched files
		result.FilesPatched = append(result.FilesPatched, targetPath)
	}

	return result, nil
}

// parsePatch parses a patch file content into a slice of FilePatch objects
func parsePatch(patchContent string) ([]FilePatch, error) {
	var patches []FilePatch
	var currentPatch *FilePatch

	// Split the patch content into lines
	lines := strings.Split(patchContent, "\n")

	// Regular expressions for parsing patch headers
	fileHeaderRegex := regexp.MustCompile(`^--- ([^\t\n]+)[\t]*.*$`)
	fileTargetRegex := regexp.MustCompile(`^\+\+\+ ([^\t\n]+)[\t]*.*$`)
	hunkHeaderRegex := regexp.MustCompile(`^@@ -(\d+),(\d+) \+(\d+),(\d+) @@.*$`)

	// Process each line
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for file header (--- line)
		if matches := fileHeaderRegex.FindStringSubmatch(line); matches != nil {
			// If we already have a patch, add it to the list
			if currentPatch != nil && currentPatch.TargetFile != "" && len(currentPatch.Hunks) > 0 {
				patches = append(patches, *currentPatch)
			}

			// Start a new file patch
			currentPatch = &FilePatch{
				SourceFile: matches[1],
				Hunks:      []Hunk{},
			}
			continue
		}

		// Check for file target (+++ line)
		if currentPatch != nil && currentPatch.TargetFile == "" {
			if matches := fileTargetRegex.FindStringSubmatch(line); matches != nil {
				currentPatch.TargetFile = matches[1]
				continue
			}
		}

		// Check for hunk header (@@ line)
		if currentPatch != nil && currentPatch.TargetFile != "" {
			if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
				// Parse hunk header
				sourceStart, _ := strconv.Atoi(matches[1])
				sourceLines, _ := strconv.Atoi(matches[2])
				targetStart, _ := strconv.Atoi(matches[3])
				targetLines, _ := strconv.Atoi(matches[4])

				// Start a new hunk
				hunk := Hunk{
					SourceStart: sourceStart,
					SourceLines: sourceLines,
					TargetStart: targetStart,
					TargetLines: targetLines,
					Context:     []string{},
					Added:       []string{},
					Removed:     []string{},
				}

				// Process hunk content
				j := i + 1
				for j < len(lines) {
					contentLine := lines[j]

					// Stop when we reach the next hunk or file
					if strings.HasPrefix(contentLine, "@@") ||
						strings.HasPrefix(contentLine, "---") ||
						strings.HasPrefix(contentLine, "+++") {
						break
					}

					// Process the line based on its prefix
					if strings.HasPrefix(contentLine, "+") {
						// Added line
						hunk.Added = append(hunk.Added, contentLine[1:])
					} else if strings.HasPrefix(contentLine, "-") {
						// Removed line
						hunk.Removed = append(hunk.Removed, contentLine[1:])
					} else if strings.HasPrefix(contentLine, " ") {
						// Context line
						hunk.Context = append(hunk.Context, contentLine[1:])
					}

					j++
				}

				// Update the line index
				i = j - 1

				// Add the hunk to the current patch
				currentPatch.Hunks = append(currentPatch.Hunks, hunk)
				continue
			}
		}
	}

	// Add the last file patch if it exists
	if currentPatch != nil && currentPatch.TargetFile != "" && len(currentPatch.Hunks) > 0 {
		patches = append(patches, *currentPatch)
	}

	return patches, nil
}

// applyStripLevel applies the strip level to a file path
func applyStripLevel(filePath string, stripLevel int) string {
	parts := strings.Split(filePath, "/")
	if stripLevel >= len(parts) {
		return parts[len(parts)-1]
	}
	return strings.Join(parts[stripLevel:], "/")
}

// applyHunks applies a list of hunks to a file content
func applyHunks(content string, hunks []Hunk) (string, int, int) {
	lines := strings.Split(content, "\n")
	hunksApplied := 0
	hunksFailed := 0

	// Apply hunks in reverse order to avoid line number changes
	for i := len(hunks) - 1; i >= 0; i-- {
		hunk := hunks[i]

		// Find the hunk location
		hunkLocation := findHunkLocation(lines, hunk)
		if hunkLocation == -1 {
			hunksFailed++
			continue
		}

		// Apply the hunk
		newLines := applyHunk(lines, hunk, hunkLocation)
		if newLines != nil {
			lines = newLines
			hunksApplied++
		} else {
			hunksFailed++
		}
	}

	return strings.Join(lines, "\n"), hunksApplied, hunksFailed
}

// findHunkLocation finds the location of a hunk in the file content
func findHunkLocation(lines []string, hunk Hunk) int {
	// Try to find the removed lines in the file
	for i := 0; i < len(lines); i++ {
		// Check if we have enough lines left to match
		if i+len(hunk.Removed) > len(lines) {
			continue
		}

		// Check if all removed lines match
		match := true
		for j := 0; j < len(hunk.Removed); j++ {
			if lines[i+j] != hunk.Removed[j] {
				match = false
				break
			}
		}

		if match {
			return i
		}
	}

	return -1
}

// applyHunk applies a hunk to the file content
func applyHunk(lines []string, hunk Hunk, location int) []string {
	// Create a new slice for the result
	result := make([]string, 0, len(lines)+len(hunk.Added)-len(hunk.Removed))

	// Copy lines before the hunk
	result = append(result, lines[:location]...)

	// Add the added lines
	result = append(result, hunk.Added...)

	// Skip the removed lines in the original content
	skipLines := len(hunk.Removed)

	// Copy lines after the hunk
	result = append(result, lines[location+skipLines:]...)

	return result
}

// RegisterPatch registers the patch tool with the MCP server
func RegisterPatch(mcpServer *server.MCPServer) {
	// Create the tool definition
	patchTool := mcp.NewTool("patch",
		mcp.WithDescription("Applies patches to files using the standard unified diff format"),
		mcp.WithString("patch_content",
			mcp.Description("The content of the patch file in unified diff format"),
			mcp.Required(),
		),
		mcp.WithString("target_directory",
			mcp.Description("The directory where the files to be patched are located (default: current directory)"),
		),
		mcp.WithNumber("strip_level",
			mcp.Description("The number of leading directories to strip from file names (default: 0)"),
		),
		mcp.WithBoolean("dry_run",
			mcp.Description("If true, show what would be done but don't actually modify any files (default: false)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("patch", HandlePatch)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(patchTool, wrappedHandler)
}
