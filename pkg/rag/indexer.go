package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// indexRepository indexes a repository for RAG
func indexRepository(repoPath string, filePatterns []string) (*IndexResult, error) {
	startTime := time.Now()

	// Validate repository path
	repoInfo, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing repository: %v", err)
	}
	if !repoInfo.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", repoPath)
	}

	// Create index directory if it doesn't exist
	indexDir := filepath.Join(repoPath, ".rag-index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating index directory: %v", err)
	}

	// Initialize result
	result := &IndexResult{
		FileTypes: make(map[string]int),
	}

	// Find files matching the patterns
	var files []string
	for _, pattern := range filePatterns {
		err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
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

			// Check if file matches the pattern
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				files = append(files, path)

				// Count file types
				ext := filepath.Ext(path)
				result.FileTypes[ext]++
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error walking repository: %v", err)
		}
	}

	// Process files
	result.FilesIndexed = len(files)
	result.SnippetsIndexed = result.FilesIndexed * 5  // Assume 5 snippets per file on average
	result.TotalTokens = result.SnippetsIndexed * 100 // Assume 100 tokens per snippet on average
	result.IndexSize = int64(result.TotalTokens * 4)  // Assume 4 bytes per token

	// In a real implementation, we would:
	// 1. Read each file
	// 2. Split into snippets
	// 3. Generate embeddings for each snippet
	// 4. Store embeddings in a vector database

	// For demonstration purposes, we'll just create a mock index file
	indexFile := filepath.Join(indexDir, "index.json")
	indexContent := fmt.Sprintf(`{
  "files_indexed": %d,
  "snippets_indexed": %d,
  "total_tokens": %d,
  "index_size": %d,
  "timestamp": "%s"
}`, result.FilesIndexed, result.SnippetsIndexed, result.TotalTokens, result.IndexSize, time.Now().Format(time.RFC3339))

	if err := os.WriteFile(indexFile, []byte(indexContent), 0644); err != nil {
		return nil, fmt.Errorf("error writing index file: %v", err)
	}

	result.TimeTaken = time.Since(startTime)
	return result, nil
}
