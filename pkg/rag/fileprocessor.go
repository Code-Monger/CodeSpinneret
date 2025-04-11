package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// processFile processes a file and extracts code snippets
func processFile(filePath string) ([]string, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Extract snippets
	snippets := extractSnippets(string(content))
	return snippets, nil
}

// extractSnippets extracts code snippets from content
func extractSnippets(content string) []string {
	// In a real implementation, we would:
	// 1. Parse the code to understand its structure
	// 2. Extract meaningful snippets (functions, classes, etc.)
	// 3. Return the snippets

	// For demonstration purposes, we'll just split the content into chunks
	lines := strings.Split(content, "\n")
	var snippets []string

	// If content is short, return the whole thing as one snippet
	if len(lines) <= 20 {
		return []string{content}
	}

	// Split into chunks of about 20 lines
	chunkSize := 20
	for i := 0; i < len(lines); i += chunkSize {
		end := min(i+chunkSize, len(lines))
		snippet := strings.Join(lines[i:end], "\n")
		snippets = append(snippets, snippet)
	}

	return snippets
}

// processDirectory processes all files in a directory
func processDirectory(dirPath string, patterns []string) ([]string, error) {
	var allSnippets []string

	// Walk the directory
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		matched := false
		for _, pattern := range patterns {
			if match, _ := filepath.Match(pattern, filepath.Base(path)); match {
				matched = true
				break
			}
		}
		if !matched {
			return nil
		}

		// Process the file
		snippets, err := processFile(path)
		if err != nil {
			return err
		}

		// Add snippets
		allSnippets = append(allSnippets, snippets...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return allSnippets, nil
}
