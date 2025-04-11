package filesearch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleFileSearch is the handler function for the file search tool
func HandleFileSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract directory path
	directory, ok := arguments["directory"].(string)
	if !ok {
		return nil, fmt.Errorf("directory must be a string")
	}

	// Extract pattern
	pattern, _ := arguments["pattern"].(string)
	// If pattern is not provided, match all files
	if pattern == "" {
		pattern = "*"
	}

	// Extract content pattern
	contentPattern, hasContentPattern := arguments["content_pattern"].(string)

	// Extract recursive flag
	recursive, _ := arguments["recursive"].(bool)

	// Extract modified after
	modifiedAfterStr, hasModifiedAfter := arguments["modified_after"].(string)
	var modifiedAfter time.Time
	if hasModifiedAfter {
		var err error
		modifiedAfter, err = time.Parse(time.RFC3339, modifiedAfterStr)
		if err != nil {
			return nil, fmt.Errorf("invalid modified_after date format, use RFC3339 (e.g., 2006-01-02T15:04:05Z07:00): %v", err)
		}
	}

	// Extract size constraints
	minSizeStr, hasMinSize := arguments["min_size"].(string)
	maxSizeStr, hasMaxSize := arguments["max_size"].(string)

	var minSize, maxSize int64
	if hasMinSize {
		var err error
		minSize, err = parseSize(minSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid min_size format: %v", err)
		}
	}

	if hasMaxSize {
		var err error
		maxSize, err = parseSize(maxSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid max_size format: %v", err)
		}
	}

	// Compile content regex if provided
	var contentRegex *regexp.Regexp
	if hasContentPattern {
		var err error
		contentRegex, err = regexp.Compile(contentPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid content_pattern regex: %v", err)
		}
	}

	// Search for files
	results, err := searchFiles(directory, pattern, contentRegex, recursive, modifiedAfter, minSize, maxSize)
	if err != nil {
		return nil, fmt.Errorf("error searching files: %v", err)
	}

	// Format results
	resultText := fmt.Sprintf("Found %d files matching criteria:\n\n", len(results))
	for _, file := range results {
		resultText += fmt.Sprintf("- %s\n", file.Path)
		resultText += fmt.Sprintf("  Size: %s\n", formatSize(file.Size))
		resultText += fmt.Sprintf("  Modified: %s\n", file.ModTime.Format(time.RFC3339))
		if file.ContentMatch != "" {
			resultText += fmt.Sprintf("  Content match: %s\n", file.ContentMatch)
		}
		resultText += "\n"
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

// FileResult represents a file search result
type FileResult struct {
	Path         string
	Size         int64
	ModTime      time.Time
	ContentMatch string
}

// searchFiles searches for files matching the given criteria
func searchFiles(directory, pattern string, contentRegex *regexp.Regexp, recursive bool, modifiedAfter time.Time, minSize, maxSize int64) ([]FileResult, error) {
	var results []FileResult

	// Validate directory
	dirInfo, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("directory error: %v", err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", directory)
	}

	// Define the walk function
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories if we're only interested in files
		if info.IsDir() {
			// If not recursive, skip subdirectories
			if !recursive && path != directory {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file matches the pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		// Check modified time
		if !modifiedAfter.IsZero() && info.ModTime().Before(modifiedAfter) {
			return nil
		}

		// Check size constraints
		if minSize > 0 && info.Size() < minSize {
			return nil
		}
		if maxSize > 0 && info.Size() > maxSize {
			return nil
		}

		// Check content if regex is provided
		var contentMatch string
		if contentRegex != nil {
			content, err := os.ReadFile(path)
			if err != nil {
				// Skip files that can't be read
				return nil
			}

			if contentRegex.Match(content) {
				// Find the first match for display
				loc := contentRegex.FindIndex(content)
				if loc != nil {
					// Get some context around the match
					start := loc[0]
					end := loc[1]

					// Get up to 50 chars before and after the match
					contextStart := start - 50
					if contextStart < 0 {
						contextStart = 0
					}
					contextEnd := end + 50
					if contextEnd > len(content) {
						contextEnd = len(content)
					}

					// Extract the context
					context := content[contextStart:contextEnd]
					// Convert to string and clean up for display
					contextStr := string(context)
					// Replace newlines with spaces for display
					contextStr = strings.ReplaceAll(contextStr, "\n", " ")
					// Truncate if too long
					if len(contextStr) > 100 {
						contextStr = contextStr[:97] + "..."
					}
					contentMatch = contextStr
				}
			} else {
				return nil // Skip files that don't match content
			}
		}

		// Add to results
		results = append(results, FileResult{
			Path:         path,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			ContentMatch: contentMatch,
		})

		return nil
	}

	// Walk the directory
	err = filepath.Walk(directory, walkFn)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// parseSize parses a size string (e.g., "10MB") into bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, nil
	}

	// Check for unit suffix
	var multiplier int64 = 1
	var numStr string

	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		numStr = sizeStr[:len(sizeStr)-2]
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		numStr = sizeStr[:len(sizeStr)-2]
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		numStr = sizeStr[:len(sizeStr)-2]
	} else if strings.HasSuffix(sizeStr, "B") {
		numStr = sizeStr[:len(sizeStr)-1]
	} else {
		// Assume bytes if no unit is specified
		numStr = sizeStr
	}

	// Parse the numeric part
	var size int64
	_, err := fmt.Sscanf(numStr, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("invalid size format: %v", err)
	}

	return size * multiplier, nil
}

// formatSize formats a size in bytes to a human-readable string
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// RegisterFileSearch registers the file search tool with the MCP server
func RegisterFileSearch(mcpServer *server.MCPServer) {
	mcpServer.AddTool(mcp.NewTool("filesearch",
		mcp.WithDescription("Search for files based on various criteria like name patterns, content, size, and modification time"),
		mcp.WithString("directory",
			mcp.Description("Directory to search in"),
			mcp.Required(),
		),
		mcp.WithString("pattern",
			mcp.Description("File name pattern (e.g., '*.go', 'file?.txt')"),
		),
		mcp.WithString("content_pattern",
			mcp.Description("Regular expression to match file content"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to search recursively in subdirectories"),
		),
		mcp.WithString("modified_after",
			mcp.Description("Only include files modified after this date (RFC3339 format)"),
		),
		mcp.WithString("min_size",
			mcp.Description("Minimum file size (e.g., '10KB', '5MB')"),
		),
		mcp.WithString("max_size",
			mcp.Description("Maximum file size (e.g., '10KB', '5MB')"),
		),
	), HandleFileSearch)
}
