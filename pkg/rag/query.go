package rag

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// queryRepository queries a repository using RAG
func queryRepository(repoPath string, query string, numResults int) (*QueryResult, error) {
	startTime := time.Now()

	// Validate repository path
	repoInfo, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing repository: %v", err)
	}
	if !repoInfo.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", repoPath)
	}

	// Check if index exists
	indexDir := filepath.Join(repoPath, ".rag-index")
	indexInfo, err := os.Stat(indexDir)
	if err != nil || !indexInfo.IsDir() {
		return nil, fmt.Errorf("repository is not indexed: %s", repoPath)
	}

	// In a real implementation, we would:
	// 1. Generate an embedding for the query
	// 2. Search the vector database for similar snippets
	// 3. Return the most similar snippets

	// For demonstration purposes, we'll just return some mock results
	result := &QueryResult{
		Results: []CodeSnippet{},
	}

	// Find some random files to use as mock results
	var files []string
	err = filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the index directory
		if info.IsDir() {
			if path == indexDir {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include source code files
		ext := filepath.Ext(path)
		if isSourceCodeFile(ext) {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking repository: %v", err)
	}

	// If no files found, return empty result
	if len(files) == 0 {
		result.TimeTaken = time.Since(startTime)
		return result, nil
	}

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate mock results
	for i := 0; i < min(numResults, len(files)); i++ {
		// Pick a random file
		fileIndex := rand.Intn(len(files))
		filePath := files[fileIndex]

		// Read the file
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		// Extract a snippet
		snippet := extractSnippet(string(content), query)

		// Add to results
		result.Results = append(result.Results, CodeSnippet{
			FilePath:   filePath,
			Snippet:    snippet,
			Similarity: 0.5 + rand.Float64()*0.5, // Random similarity between 0.5 and 1.0
		})

		// Remove the file from the list to avoid duplicates
		files = append(files[:fileIndex], files[fileIndex+1:]...)
	}

	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// extractSnippet extracts a relevant snippet from the content
func extractSnippet(content, query string) string {
	lines := strings.Split(content, "\n")

	// If content is short, return the whole thing
	if len(lines) <= 10 {
		return content
	}

	// Try to find a line containing the query
	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
			// Return a snippet around this line
			start := max(0, i-5)
			end := min(len(lines), i+6)
			return strings.Join(lines[start:end], "\n")
		}
	}

	// If query not found, return the first 10 lines
	return strings.Join(lines[:10], "\n")
}

// isSourceCodeFile checks if a file extension is for a source code file
func isSourceCodeFile(ext string) bool {
	sourceExts := map[string]bool{
		".go":   true,
		".js":   true,
		".ts":   true,
		".py":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".cs":   true,
		".html": true,
		".css":  true,
	}
	return sourceExts[ext]
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
