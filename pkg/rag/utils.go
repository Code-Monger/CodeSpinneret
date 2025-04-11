package rag

import (
	"fmt"
	"path/filepath"
	"strings"
)

// generateResponse generates a response based on the query and retrieved snippets
func generateResponse(query string, snippets []CodeSnippet) string {
	if len(snippets) == 0 {
		return "I couldn't find any relevant code snippets for your query."
	}

	// In a real implementation, we would:
	// 1. Use an LLM to generate a response based on the query and snippets
	// 2. Return the generated response

	// For demonstration purposes, we'll just return a mock response
	response := fmt.Sprintf("Based on your query '%s', I found the following information:\n\n", query)

	// Add information about the most relevant snippet
	response += fmt.Sprintf("The most relevant code is in file '%s'. ", snippets[0].FilePath)

	// Add information about the language
	language := getLanguageFromFilePath(snippets[0].FilePath)
	response += fmt.Sprintf("This is %s code. ", language)

	// Add a simple analysis of the snippet
	lines := strings.Split(snippets[0].Snippet, "\n")
	response += fmt.Sprintf("The snippet is %d lines long. ", len(lines))

	// Add a simple recommendation
	response += "\n\nBased on this code, I recommend reviewing it for clarity and potential optimizations."

	return response
}

// getLanguageFromFilePath returns the programming language based on the file extension
func getLanguageFromFilePath(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".py":
		return "python"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".md":
		return "markdown"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".yaml", ".yml":
		return "yaml"
	case ".sh":
		return "bash"
	case ".bat", ".cmd":
		return "batch"
	case ".ps1":
		return "powershell"
	default:
		return "text"
	}
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
