package searchreplace

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleSearchReplace is the handler function for the search replace tool
func HandleSearchReplace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract directory path
	directory, ok := arguments["directory"].(string)
	if !ok {
		return nil, fmt.Errorf("directory must be a string")
	}

	// Extract file pattern
	filePattern, ok := arguments["file_pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("file_pattern must be a string")
	}

	// Extract search pattern
	searchPattern, ok := arguments["search_pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("search_pattern must be a string")
	}

	// Extract replacement text
	replacement, ok := arguments["replacement"].(string)
	if !ok {
		return nil, fmt.Errorf("replacement must be a string")
	}

	// Extract regex flag
	useRegex, _ := arguments["use_regex"].(bool)

	// Extract recursive flag
	recursive, _ := arguments["recursive"].(bool)

	// Extract preview flag
	preview, _ := arguments["preview"].(bool)

	// Extract case sensitive flag
	caseSensitive, _ := arguments["case_sensitive"].(bool)

	// Perform search and replace
	results, err := searchAndReplace(directory, filePattern, searchPattern, replacement, useRegex, recursive, preview, caseSensitive)
	if err != nil {
		return nil, fmt.Errorf("error performing search and replace: %v", err)
	}

	// Format results
	resultText := fmt.Sprintf("Search and Replace Results:\n\n")
	if preview {
		resultText += "Preview mode: No files were modified\n\n"
	}

	resultText += fmt.Sprintf("Files processed: %d\n", results.FilesProcessed)
	resultText += fmt.Sprintf("Files modified: %d\n", results.FilesModified)
	resultText += fmt.Sprintf("Total replacements: %d\n\n", results.TotalReplacements)

	if len(results.Details) > 0 {
		resultText += "Details:\n"
		for _, detail := range results.Details {
			resultText += fmt.Sprintf("- %s: %d replacements\n", detail.FilePath, detail.Replacements)

			// Add sample of changes if available
			if detail.Sample != "" {
				resultText += fmt.Sprintf("  Sample: %s\n", detail.Sample)
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// SearchReplaceResults represents the results of a search and replace operation
type SearchReplaceResults struct {
	FilesProcessed    int
	FilesModified     int
	TotalReplacements int
	Details           []FileDetail
}

// FileDetail represents details about replacements in a specific file
type FileDetail struct {
	FilePath     string
	Replacements int
	Sample       string
}

// searchAndReplace performs search and replace operations on files
func searchAndReplace(directory, filePattern, searchPattern, replacement string, useRegex, recursive, preview, caseSensitive bool) (*SearchReplaceResults, error) {
	results := &SearchReplaceResults{
		Details: []FileDetail{},
	}

	// Validate directory
	dirInfo, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("directory error: %v", err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", directory)
	}

	// Check if file pattern is valid
	_, err = filepath.Match(filePattern, "")
	if err != nil {
		return nil, fmt.Errorf("invalid file pattern: %v", err)
	}

	// Prepare search pattern
	var searchRegex *regexp.Regexp
	if useRegex {
		if caseSensitive {
			searchRegex, err = regexp.Compile(searchPattern)
		} else {
			searchRegex, err = regexp.Compile("(?i)" + searchPattern)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid search pattern regex: %v", err)
		}
	} else if !caseSensitive {
		searchPattern = strings.ToLower(searchPattern)
	}

	// Walk through the directory
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// If not recursive, skip subdirectories
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches the pattern
		matched, err := filepath.Match(filePattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		// Process the file
		fileDetail, err := processFile(path, searchPattern, replacement, searchRegex, preview, caseSensitive)
		if err != nil {
			return err
		}

		// Update results
		results.FilesProcessed++
		if fileDetail.Replacements > 0 {
			results.FilesModified++
			results.TotalReplacements += fileDetail.Replacements
			results.Details = append(results.Details, *fileDetail)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// processFile performs search and replace on a single file
func processFile(filePath, searchPattern, replacement string, searchRegex *regexp.Regexp, preview, caseSensitive bool) (*FileDetail, error) {
	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	fileDetail := &FileDetail{
		FilePath: filePath,
	}

	// Convert content to string
	contentStr := string(content)
	originalContent := contentStr

	// Perform replacement
	var newContent string
	if searchRegex != nil {
		// Use regex replacement
		newContent = searchRegex.ReplaceAllString(contentStr, replacement)
		fileDetail.Replacements = len(searchRegex.FindAllString(contentStr, -1))
	} else {
		// Use simple string replacement
		if caseSensitive {
			newContent = strings.ReplaceAll(contentStr, searchPattern, replacement)
			fileDetail.Replacements = strings.Count(contentStr, searchPattern)
		} else {
			// Case insensitive replacement without regex
			lowerContent := strings.ToLower(contentStr)
			lastIndex := 0
			count := 0
			newContent = contentStr

			for {
				index := strings.Index(lowerContent[lastIndex:], searchPattern)
				if index == -1 {
					break
				}

				// Adjust index to account for the slice
				index += lastIndex

				// Replace in the original case
				newContent = newContent[:index] + replacement + newContent[index+len(searchPattern):]

				// Update the lowercase content to match
				lowerContent = strings.ToLower(newContent)

				// Move past this replacement
				lastIndex = index + len(replacement)
				count++
			}

			fileDetail.Replacements = count
		}
	}

	// If there were replacements, add a sample
	if fileDetail.Replacements > 0 {
		// Find a context for the first replacement
		var beforeContext, afterContext string
		var replacementIndex int

		if searchRegex != nil {
			loc := searchRegex.FindStringIndex(originalContent)
			if loc != nil {
				replacementIndex = loc[0]
			}
		} else if caseSensitive {
			replacementIndex = strings.Index(originalContent, searchPattern)
		} else {
			lowerContent := strings.ToLower(originalContent)
			replacementIndex = strings.Index(lowerContent, searchPattern)
		}

		if replacementIndex >= 0 {
			// Get context before replacement
			startContext := replacementIndex - 20
			if startContext < 0 {
				startContext = 0
			}
			beforeContext = originalContent[startContext:replacementIndex]

			// Get context after replacement
			endContext := replacementIndex + len(searchPattern) + 20
			if endContext > len(originalContent) {
				endContext = len(originalContent)
			}
			afterContext = originalContent[replacementIndex+len(searchPattern) : endContext]

			// Create sample
			fileDetail.Sample = fmt.Sprintf("...%s[%s -> %s]%s...",
				beforeContext,
				originalContent[replacementIndex:replacementIndex+len(searchPattern)],
				replacement,
				afterContext)
		}

		// Write the changes to the file if not in preview mode
		if !preview {
			// Get file info to preserve file mode
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				return nil, fmt.Errorf("error getting file info for %s: %v", filePath, err)
			}

			// Write the changes to the file
			err = ioutil.WriteFile(filePath, []byte(newContent), fileInfo.Mode())
			if err != nil {
				return nil, fmt.Errorf("error writing to file %s: %v", filePath, err)
			}
		}
	}

	return fileDetail, nil
}

// RegisterSearchReplace registers the search replace tool with the MCP server
func RegisterSearchReplace(mcpServer *server.MCPServer) {
	mcpServer.AddTool(mcp.NewTool("searchreplace",
		mcp.WithDescription("Find and replace text in files, with support for regular expressions and batch operations"),
		mcp.WithString("directory",
			mcp.Description("Directory to search in"),
			mcp.Required(),
		),
		mcp.WithString("file_pattern",
			mcp.Description("File name pattern (e.g., '*.txt', '*.go')"),
			mcp.Required(),
		),
		mcp.WithString("search_pattern",
			mcp.Description("Text or pattern to search for"),
			mcp.Required(),
		),
		mcp.WithString("replacement",
			mcp.Description("Text to replace with"),
			mcp.Required(),
		),
		mcp.WithBoolean("use_regex",
			mcp.Description("Whether to use regular expressions for search pattern"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to search recursively in subdirectories"),
		),
		mcp.WithBoolean("preview",
			mcp.Description("Preview changes without modifying files"),
		),
		mcp.WithBoolean("case_sensitive",
			mcp.Description("Whether the search should be case sensitive"),
		),
	), HandleSearchReplace)
}
