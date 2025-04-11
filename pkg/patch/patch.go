package patch

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/Code-Monger/CodeSpinneret/pkg/workspace"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Environment variable name for patch root directory
const EnvPatchRootDir = "PATCH_ROOT_DIR"

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

	// Extract session ID
	sessionID, _ := arguments["session_id"].(string)

	// Get root directory from workspace
	rootDir := workspace.GetRootDir(sessionID)
	if rootDir == "" {
		rootDir = targetDir // Default to target directory if workspace not set
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
	result, err := applyPatch(patchContent, targetDir, rootDir, stripLevel, dryRun)
	if err != nil {
		return nil, fmt.Errorf("error applying patch: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Patch Application Results:\n\n")
	resultText += fmt.Sprintf("Target directory: %s\n", targetDir)
	resultText += fmt.Sprintf("Root directory: %s\n", rootDir)
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
func applyPatch(patchContent, targetDir, rootDir string, stripLevel int, dryRun bool) (*PatchResult, error) {
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

		// Resolve the path relative to the root directory
		relPath, err := filepath.Rel(rootDir, filepath.Join(rootDir, targetPath))
		if err != nil {
			result.FilesSkipped[targetPath] = fmt.Sprintf("error resolving relative path: %v", err)
			continue
		}

		// Resolve the full path in the target directory
		fullPath := filepath.Join(targetDir, relPath)

		// Check if the file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			result.FilesSkipped[targetPath] = "file does not exist"
			continue
		}

		// Read the file content
		content, err := os.ReadFile(fullPath)
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
			err = os.WriteFile(fullPath, []byte(newContent), 0644)
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

				// Keep track of context lines
				var contextLines []string

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
						contextLine := contentLine[1:]
						contextLines = append(contextLines, contextLine)
					}

					j++
				}

				// Set all context lines
				hunk.Context = contextLines

				// Set the context lines
				hunk.Context = contextLines

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
	// If there are no removed lines, we can't find the location
	if len(hunk.Removed) == 0 {
		// If there are context lines, try to find them
		if len(hunk.Context) > 0 {
			return findContextLocation(lines, hunk.Context)
		}
		return -1
	}

	// First try to find the exact match with context lines
	if len(hunk.Context) > 0 {
		// Create a pattern with context lines and removed lines
		pattern := make([]string, 0, len(hunk.Context)+len(hunk.Removed))

		// Add all context and removed lines in the correct order
		contextIndex := 0
		removedIndex := 0
		inRemovedSection := false

		// Process context lines that come before removed lines
		for contextIndex < len(hunk.Context) && !inRemovedSection {
			// Check if this context line is actually a removed line
			if removedIndex < len(hunk.Removed) && hunk.Context[contextIndex] == hunk.Removed[removedIndex] {
				inRemovedSection = true
				break
			}

			pattern = append(pattern, hunk.Context[contextIndex])
			contextIndex++
		}

		// Add all removed lines
		pattern = append(pattern, hunk.Removed...)

		// Add remaining context lines that come after removed lines
		for contextIndex < len(hunk.Context) {
			pattern = append(pattern, hunk.Context[contextIndex])
			contextIndex++
		}

		// Try to find the pattern in the file
		for i := 0; i <= len(lines)-len(pattern); i++ {
			match := true
			for j := 0; j < len(pattern); j++ {
				if i+j >= len(lines) || lines[i+j] != pattern[j] {
					match = false
					break
				}
			}

			if match {
				// Return the position where removed lines start
				return i + (len(pattern) - len(hunk.Removed) - (len(hunk.Context) - contextIndex))
			}
		}
	}

	// If we couldn't find the pattern with context, try just the removed lines
	for i := 0; i <= len(lines)-len(hunk.Removed); i++ {
		match := true
		for j := 0; j < len(hunk.Removed); j++ {
			if i+j >= len(lines) || lines[i+j] != hunk.Removed[j] {
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

// findContextLocation finds the location of context lines in the file content
func findContextLocation(lines []string, context []string) int {
	for i := 0; i <= len(lines)-len(context); i++ {
		match := true
		for j := 0; j < len(context); j++ {
			if i+j >= len(lines) || lines[i+j] != context[j] {
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
		mcp.WithDescription("Applies patches to files using the standard unified diff format. Supports both file-specific and directory-wide patching with configurable path handling. Uses the workspace session for consistent relative path resolution across tools. Provides detailed reporting of applied and failed hunks, with dry-run capability for safe testing. Ideal for code modifications, bug fixes, and implementing changes from external sources."),
		mcp.WithString("patch_content",
			mcp.Description("The content of the patch file in unified diff format, containing the changes to apply to one or more files"),
			mcp.Required(),
		),
		mcp.WithString("target_directory",
			mcp.Description("The directory where the files to be patched are located (absolute or relative path, default: current directory)"),
		),
		mcp.WithString("session_id",
			mcp.Description("Session ID to use for resolving relative paths"),
		),
		mcp.WithNumber("strip_level",
			mcp.Description("The number of leading directories to strip from file paths in the patch, useful for applying patches created in different directory structures (default: 0)"),
		),
		mcp.WithBoolean("dry_run",
			mcp.Description("If true, performs a simulation showing what would be changed without actually modifying any files, useful for testing patches before applying them (default: false)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("patch", HandlePatch)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(patchTool, wrappedHandler)

	// Log the registration
	log.Printf("[Patch] Registered patch tool")
}
