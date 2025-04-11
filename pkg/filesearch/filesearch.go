package filesearch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
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

	// Extract file name pattern
	pattern, _ := arguments["pattern"].(string)

	// Extract content pattern
	contentPattern, _ := arguments["content_pattern"].(string)

	// Extract recursive flag
	recursive := false
	if recursiveBool, ok := arguments["recursive"].(bool); ok {
		recursive = recursiveBool
	}

	// Extract modified after date
	var modifiedAfter time.Time
	if modifiedAfterStr, ok := arguments["modified_after"].(string); ok && modifiedAfterStr != "" {
		var err error
		modifiedAfter, err = time.Parse(time.RFC3339, modifiedAfterStr)
		if err != nil {
			return nil, fmt.Errorf("invalid modified_after date format: %v", err)
		}
	}

	// Extract min size
	var minSize int64
	if minSizeStr, ok := arguments["min_size"].(string); ok && minSizeStr != "" {
		var err error
		minSize, err = parseSize(minSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid min_size format: %v", err)
		}
	}

	// Extract max size
	var maxSize int64
	if maxSizeStr, ok := arguments["max_size"].(string); ok && maxSizeStr != "" {
		var err error
		maxSize, err = parseSize(maxSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid max_size format: %v", err)
		}
	}

	// Compile content pattern regex if provided
	var contentRegex *regexp.Regexp
	if contentPattern != "" {
		var err error
		contentRegex, err = regexp.Compile(contentPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid content pattern regex: %v", err)
		}
	}

	// Search for files
	results, err := searchFiles(directory, pattern, contentRegex, recursive, modifiedAfter, minSize, maxSize)
	if err != nil {
		return nil, fmt.Errorf("error searching files: %v", err)
	}

	// Format the results
	resultText := fmt.Sprintf("Found %d files matching the criteria:\n\n", len(results))
	for _, result := range results {
		resultText += fmt.Sprintf("Path: %s\n", result.Path)
		resultText += fmt.Sprintf("Size: %s\n", formatSize(result.Size))
		resultText += fmt.Sprintf("Modified: %s\n", result.ModTime.Format(time.RFC3339))
		if result.ContentMatch != "" {
			resultText += fmt.Sprintf("Content match: %s\n", result.ContentMatch)
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

	// Ensure the directory exists
	info, err := os.Stat(directory)
	if err != nil {
		return nil, fmt.Errorf("error accessing directory: %v", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", directory)
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

		// Check file name pattern
		if pattern != "" {
			matched, err := filepath.Match(pattern, info.Name())
			if err != nil {
				return err
			}
			if !matched {
				return nil
			}
		}

		// Check modified time
		if !modifiedAfter.IsZero() && info.ModTime().Before(modifiedAfter) {
			return nil
		}

		// Check file size
		if minSize > 0 && info.Size() < minSize {
			return nil
		}
		if maxSize > 0 && info.Size() > maxSize {
			return nil
		}

		// Check file content if needed
		var contentMatch string
		if contentRegex != nil {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}

			if matches := contentRegex.FindSubmatch(content); len(matches) > 0 {
				contentMatch = string(matches[0])
			} else {
				return nil
			}
		}

		// Add the file to the results
		results = append(results, FileResult{
			Path:         path,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			ContentMatch: contentMatch,
		})

		return nil
	}

	err = filepath.Walk(directory, walkFn)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// parseSize parses a human-readable size string (e.g., "10KB", "5MB") to bytes
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" {
		return 0, nil
	}

	// Define size multipliers
	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}

	// Default multiplier is bytes
	multiplier := int64(1)
	numStr := sizeStr

	// Check for unit suffix
	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = multipliers["KB"]
		numStr = sizeStr[:len(sizeStr)-2]
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = multipliers["MB"]
		numStr = sizeStr[:len(sizeStr)-2]
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = multipliers["GB"]
		numStr = sizeStr[:len(sizeStr)-2]
	} else if strings.HasSuffix(sizeStr, "TB") {
		multiplier = multipliers["TB"]
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
	// Create the tool definition
	fileSearchTool := mcp.NewTool("filesearch",
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
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("filesearch", HandleFileSearch)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(fileSearchTool, wrappedHandler)
}
