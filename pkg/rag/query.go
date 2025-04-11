package rag

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	// Create a result structure
	result := &QueryResult{
		Results: []CodeSnippet{},
	}

	// Normalize the query for better matching
	normalizedQuery := strings.ToLower(query)
	queryTerms := strings.Fields(normalizedQuery)

	// Structure to hold file matches
	type FileMatch struct {
		Path      string
		Content   string
		Score     float64
		MatchInfo string
	}

	var matches []FileMatch

	// Walk through the repository to find matching files
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
		if !isSourceCodeFile(ext) {
			return nil
		}

		// Calculate a base score based on the file path
		relPath, _ := filepath.Rel(repoPath, path)
		pathScore := calculatePathScore(relPath, normalizedQuery, queryTerms)

		// Read the file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip this file if we can't read it
		}

		contentStr := string(content)
		contentScore := calculateContentScore(contentStr, normalizedQuery, queryTerms)

		// Calculate total score
		totalScore := (pathScore + contentScore) / 2.0

		// Only include files with a minimum score
		if totalScore > 0.1 {
			matchInfo := fmt.Sprintf("Path score: %.2f, Content score: %.2f", pathScore, contentScore)
			matches = append(matches, FileMatch{
				Path:      path,
				Content:   contentStr,
				Score:     totalScore,
				MatchInfo: matchInfo,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking repository: %v", err)
	}

	// If no files found, return empty result
	if len(matches) == 0 {
		result.TimeTaken = time.Since(startTime)
		return result, nil
	}

	// Sort matches by score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Take the top N results
	for i := 0; i < min(numResults, len(matches)); i++ {
		match := matches[i]

		// Extract a relevant snippet
		snippet := extractSnippet(match.Content, query)

		// Add to results
		result.Results = append(result.Results, CodeSnippet{
			FilePath:   match.Path,
			Snippet:    snippet,
			Similarity: match.Score,
		})
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

// calculatePathScore calculates a score based on how well the file path matches the query
func calculatePathScore(path string, normalizedQuery string, queryTerms []string) float64 {
	// Normalize the path
	normalizedPath := strings.ToLower(path)

	// Direct match with the full query
	if strings.Contains(normalizedPath, normalizedQuery) {
		return 1.0
	}

	// Check for matches with individual query terms
	matchCount := 0
	for _, term := range queryTerms {
		if len(term) < 3 {
			continue // Skip very short terms
		}
		if strings.Contains(normalizedPath, term) {
			matchCount++
		}
	}

	// Calculate score based on the percentage of query terms found in the path
	if len(queryTerms) > 0 {
		return float64(matchCount) / float64(len(queryTerms))
	}

	return 0.0
}

// calculateContentScore calculates a score based on how well the file content matches the query
func calculateContentScore(content string, normalizedQuery string, queryTerms []string) float64 {
	// Normalize the content
	normalizedContent := strings.ToLower(content)

	// Count occurrences of the full query
	fullQueryCount := strings.Count(normalizedContent, normalizedQuery)

	// If the full query appears multiple times, this is a strong match
	if fullQueryCount > 0 {
		// Scale the score based on the number of occurrences, with diminishing returns
		return minFloat(1.0, 0.7+0.3*float64(min(fullQueryCount, 10))/10.0)
	}

	// Count occurrences of individual query terms
	termMatches := 0
	totalTerms := len(queryTerms)

	if totalTerms == 0 {
		return 0.0
	}

	for _, term := range queryTerms {
		// Skip very short terms (less than 3 characters)
		if len(term) < 3 {
			totalTerms--
			continue
		}

		// Count occurrences of this term
		termCount := strings.Count(normalizedContent, term)
		if termCount > 0 {
			termMatches++
		}
	}

	// Adjust for the case where all terms were too short
	if totalTerms == 0 {
		return 0.0
	}

	// Calculate score based on the percentage of query terms found in the content
	return float64(termMatches) / float64(totalTerms)
}

// minFloat returns the minimum of two float64 values
func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
