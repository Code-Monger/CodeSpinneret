package linecount

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Environment variable name for root directory (same as patch tool)
const EnvRootDir = "PATCH_ROOT_DIR"

// HandleLineCount is the handler function for the linecount tool
func HandleLineCount(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract file path
	filePath, ok := arguments["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path must be a string")
	}

	// Extract count options
	countLines := true
	if countLinesVal, ok := arguments["count_lines"].(bool); ok {
		countLines = countLinesVal
	}

	countWords := false
	if countWordsVal, ok := arguments["count_words"].(bool); ok {
		countWords = countWordsVal
	}

	countChars := false
	if countCharsVal, ok := arguments["count_chars"].(bool); ok {
		countChars = countCharsVal
	}

	// Check if the file path is absolute
	var fullPath string
	if filepath.IsAbs(filePath) {
		fullPath = filePath
	} else {
		// Get root directory from environment variable
		rootDir := os.Getenv(EnvRootDir)
		if rootDir == "" {
			rootDir = "." // Default to current directory if env var not set
		}

		// Join the root directory and file path
		fullPath = filepath.Join(rootDir, filePath)
	}

	// Log the paths for debugging
	log.Printf("[LineCount] File path: %s", filePath)
	log.Printf("[LineCount] Full path: %s", fullPath)

	// Count lines, words, and characters
	result, err := countFileStats(fullPath, countLines, countWords, countChars)
	if err != nil {
		return nil, fmt.Errorf("error counting file stats: %v", err)
	}
	// Format the result
	resultText := fmt.Sprintf("File: %s\n\n", filePath)
	resultText += fmt.Sprintf("Full path: %s\n\n", fullPath)

	if countLines {
		resultText += fmt.Sprintf("Lines: %d\n", result.Lines)
	}
	if countWords {
		resultText += fmt.Sprintf("Words: %d\n", result.Words)
	}
	if countChars {
		resultText += fmt.Sprintf("Characters: %d\n", result.Chars)
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

// FileStats represents the statistics of a file
type FileStats struct {
	Lines int
	Words int
	Chars int
}

// countFileStats counts the lines, words, and characters in a file
func countFileStats(filePath string, countLines, countWords, countChars bool) (*FileStats, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Initialize counters
	stats := &FileStats{}

	// Count lines, words, and characters
	for scanner.Scan() {
		line := scanner.Text()

		if countLines {
			stats.Lines++
		}

		if countWords {
			words := strings.Fields(line)
			stats.Words += len(words)
		}

		if countChars {
			stats.Chars += len(line)
			// Add 1 for the newline character that was stripped by scanner.Text()
			if line != "" {
				stats.Chars++
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return stats, nil
}

// RegisterLineCount registers the linecount tool with the MCP server
func RegisterLineCount(mcpServer *server.MCPServer) {
	// Create the tool definition
	lineCountTool := mcp.NewTool("linecount",
		mcp.WithDescription("Counts lines, words, and characters in a file, similar to the Unix 'wc' command. Provides detailed statistics about file content with configurable counting options. Supports both absolute and relative file paths, with automatic resolution using the PATCH_ROOT_DIR environment variable for consistent path handling across tools. Useful for code analysis, documentation metrics, and content evaluation."),
		mcp.WithString("file_path",
			mcp.Description("The path of the file to count (absolute or relative to working directory)"),
			mcp.Required(),
		),
		mcp.WithBoolean("count_lines",
			mcp.Description("Whether to count the number of lines in the file (default: true)"),
		),
		mcp.WithBoolean("count_words",
			mcp.Description("Whether to count the number of words in the file, defined as space-separated sequences of characters (default: false)"),
		),
		mcp.WithBoolean("count_chars",
			mcp.Description("Whether to count the total number of characters in the file, including whitespace and newlines (default: false)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("linecount", HandleLineCount)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(lineCountTool, wrappedHandler)

	log.Printf("[LineCount] Registered linecount tool")
}
