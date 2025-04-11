package searchreplace

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
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

	// Extract replacement
	replacement, ok := arguments["replacement"].(string)
	if !ok {
		return nil, fmt.Errorf("replacement must be a string")
	}

	// Extract use_regex flag
	useRegex := false
	if useRegexBool, ok := arguments["use_regex"].(bool); ok {
		useRegex = useRegexBool
	}

	// Extract recursive flag
	recursive := false
	if recursiveBool, ok := arguments["recursive"].(bool); ok {
		recursive = recursiveBool
	}

	// Extract preview flag
	preview := false
	if previewBool, ok := arguments["preview"].(bool); ok {
		preview = previewBool
	}

	// Extract case_sensitive flag
	caseSensitive := true
	if caseSensitiveBool, ok := arguments["case_sensitive"].(bool); ok {
		caseSensitive = caseSensitiveBool
	}

	// Perform the search and replace
	result, err := searchAndReplace(directory, filePattern, searchPattern, replacement, useRegex, recursive, preview, caseSensitive)
	if err != nil {
		return nil, fmt.Errorf("error performing search and replace: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Search and Replace Results:\n\n")
	resultText += fmt.Sprintf("Directory: %s\n", directory)
	resultText += fmt.Sprintf("File Pattern: %s\n", filePattern)
	resultText += fmt.Sprintf("Search Pattern: %s\n", searchPattern)
	resultText += fmt.Sprintf("Replacement: %s\n", replacement)
	resultText += fmt.Sprintf("Use Regex: %t\n", useRegex)
	resultText += fmt.Sprintf("Recursive: %t\n", recursive)
	resultText += fmt.Sprintf("Preview: %t\n", preview)
	resultText += fmt.Sprintf("Case Sensitive: %t\n\n", caseSensitive)

	resultText += fmt.Sprintf("Files Processed: %d\n", result.FilesProcessed)
	resultText += fmt.Sprintf("Files Modified: %d\n", result.FilesModified)
	resultText += fmt.Sprintf("Total Replacements: %d\n\n", result.TotalReplacements)

	if len(result.FileDetails) > 0 {
		resultText += "File Details:\n"
		for _, fileDetail := range result.FileDetails {
			resultText += fmt.Sprintf("\nFile: %s\n", fileDetail.FilePath)
			resultText += fmt.Sprintf("Replacements: %d\n", fileDetail.Replacements)

			if len(fileDetail.Matches) > 0 {
				resultText += "Matches:\n"
				for _, match := range fileDetail.Matches {
					resultText += fmt.Sprintf("- Line %d: %s\n", match.LineNumber, match.LineContent)
				}
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

// SearchReplaceResult represents the result of a search and replace operation
type SearchReplaceResult struct {
	FilesProcessed    int
	FilesModified     int
	TotalReplacements int
	FileDetails       []FileDetail
}

// FileDetail represents details about a file that was processed
type FileDetail struct {
	FilePath     string
	Replacements int
	Matches      []Match
}

// Match represents a match in a file
type Match struct {
	LineNumber  int
	LineContent string
}

// searchAndReplace performs a search and replace operation on files
func searchAndReplace(directory, filePattern, searchPattern, replacement string, useRegex, recursive, preview, caseSensitive bool) (*SearchReplaceResult, error) {
	result := &SearchReplaceResult{
		FileDetails: []FileDetail{},
	}

	// Compile the search pattern if using regex
	var searchRegex *regexp.Regexp
	var err error
	if useRegex {
		if caseSensitive {
			searchRegex, err = regexp.Compile(searchPattern)
		} else {
			searchRegex, err = regexp.Compile("(?i)" + searchPattern)
		}
		if err != nil {
			return nil, fmt.Errorf("invalid search pattern regex: %v", err)
		}
	}

	// Walk the directory tree
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories unless we're at the root
		if info.IsDir() {
			if path != directory && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file matches the pattern
		matched, err := filepath.Match(filePattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		// Process the file
		fileDetail, err := processFile(path, searchPattern, replacement, useRegex, preview, caseSensitive, searchRegex)
		if err != nil {
			return err
		}

		// Update the result
		result.FilesProcessed++
		if fileDetail.Replacements > 0 {
			result.FilesModified++
			result.TotalReplacements += fileDetail.Replacements
			result.FileDetails = append(result.FileDetails, *fileDetail)
		}

		return nil
	}

	err = filepath.Walk(directory, walkFn)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// processFile processes a single file for search and replace
func processFile(filePath, searchPattern, replacement string, useRegex, preview, caseSensitive bool, searchRegex *regexp.Regexp) (*FileDetail, error) {
	fileDetail := &FileDetail{
		FilePath: filePath,
		Matches:  []Match{},
	}

	// Read the file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("error accessing file %s: %v", filePath, err)
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Process the content
	var newContent string
	if useRegex {
		// Use regex for search and replace
		newContent = searchRegex.ReplaceAllString(string(content), replacement)
		fileDetail.Replacements = len(searchRegex.FindAllString(string(content), -1))
	} else {
		// Use simple string search and replace
		if caseSensitive {
			fileDetail.Replacements = strings.Count(string(content), searchPattern)
			newContent = strings.ReplaceAll(string(content), searchPattern, replacement)
		} else {
			// Case-insensitive search and replace
			lowerContent := strings.ToLower(string(content))
			lowerPattern := strings.ToLower(searchPattern)
			fileDetail.Replacements = strings.Count(lowerContent, lowerPattern)

			// Perform the replacement
			newContent = string(content)
			lastIndex := 0
			for {
				index := strings.Index(strings.ToLower(newContent[lastIndex:]), lowerPattern)
				if index == -1 {
					break
				}
				index += lastIndex
				newContent = newContent[:index] + replacement + newContent[index+len(searchPattern):]
				lastIndex = index + len(replacement)
			}
		}
	}

	// Find matches for reporting
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		var matches bool
		if useRegex {
			matches = searchRegex.MatchString(line)
		} else if caseSensitive {
			matches = strings.Contains(line, searchPattern)
		} else {
			matches = strings.Contains(strings.ToLower(line), strings.ToLower(searchPattern))
		}

		if matches {
			fileDetail.Matches = append(fileDetail.Matches, Match{
				LineNumber:  i + 1,
				LineContent: line,
			})
		}
	}

	// Write the changes to the file if not in preview mode
	if !preview && fileDetail.Replacements > 0 {
		if string(content) != newContent {
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
	// Create the tool definition
	searchReplaceTool := mcp.NewTool("searchreplace",
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
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("searchreplace", HandleSearchReplace)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(searchReplaceTool, wrappedHandler)

	// Log the registration
	log.Printf("[SearchReplace] Registered searchreplace tool")
}
