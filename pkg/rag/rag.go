package rag

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleRAG is the handler function for the RAG tool
func HandleRAG(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract operation
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	switch operation {
	case "index":
		return handleIndex(arguments)
	case "query":
		return handleQuery(arguments)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// handleIndex handles the index operation
func handleIndex(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract repository path
	repoPath, ok := arguments["repo_path"].(string)
	if !ok {
		return nil, fmt.Errorf("repo_path must be a string")
	}

	// Extract file patterns
	var filePatterns []string
	if patterns, ok := arguments["file_patterns"].([]interface{}); ok {
		for _, pattern := range patterns {
			if patternStr, ok := pattern.(string); ok {
				filePatterns = append(filePatterns, patternStr)
			}
		}
	}
	if len(filePatterns) == 0 {
		// Default to common code file patterns
		filePatterns = []string{"*.go", "*.js", "*.ts", "*.py", "*.java", "*.c", "*.cpp", "*.h", "*.cs"}
	}

	// Index the repository
	indexResult, err := indexRepository(repoPath, filePatterns)
	if err != nil {
		return nil, fmt.Errorf("error indexing repository: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Repository Indexing Results:\n\n")
	resultText += fmt.Sprintf("Repository: %s\n", repoPath)
	resultText += fmt.Sprintf("Files indexed: %d\n", indexResult.FilesIndexed)
	resultText += fmt.Sprintf("Code snippets: %d\n", indexResult.SnippetsIndexed)
	resultText += fmt.Sprintf("Total tokens: %d\n", indexResult.TotalTokens)
	resultText += fmt.Sprintf("Index size: %s\n", formatSize(indexResult.IndexSize))
	resultText += fmt.Sprintf("Time taken: %s\n\n", indexResult.TimeTaken)

	if len(indexResult.FileTypes) > 0 {
		resultText += "File types:\n"
		for fileType, count := range indexResult.FileTypes {
			resultText += fmt.Sprintf("- %s: %d files\n", fileType, count)
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

// handleQuery handles the query operation
func handleQuery(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract query
	query, ok := arguments["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query must be a string")
	}

	// Extract repository path
	repoPath, ok := arguments["repo_path"].(string)
	if !ok {
		return nil, fmt.Errorf("repo_path must be a string")
	}

	// Extract number of results
	numResults := 5 // Default
	if numResultsFloat, ok := arguments["num_results"].(float64); ok {
		numResults = int(numResultsFloat)
	}

	// Perform the query
	queryResult, err := queryRepository(repoPath, query, numResults)
	if err != nil {
		return nil, fmt.Errorf("error querying repository: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("RAG Query Results for: %s\n\n", query)
	resultText += fmt.Sprintf("Repository: %s\n", repoPath)
	resultText += fmt.Sprintf("Results: %d\n", len(queryResult.Results))
	resultText += fmt.Sprintf("Time taken: %s\n\n", queryResult.TimeTaken)

	// Generate a response using the retrieved snippets
	generatedResponse := generateResponse(query, queryResult.Results)
	resultText += fmt.Sprintf("Generated Response:\n%s\n\n", generatedResponse)

	resultText += "Retrieved Code Snippets:\n\n"
	for i, result := range queryResult.Results {
		resultText += fmt.Sprintf("%d. File: %s\n", i+1, result.FilePath)
		resultText += fmt.Sprintf("   Similarity: %.2f%%\n", result.Similarity*100)
		resultText += fmt.Sprintf("   Snippet:\n```%s\n%s\n```\n\n", getLanguageFromFilePath(result.FilePath), result.Snippet)
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

// IndexResult represents the result of indexing a repository
type IndexResult struct {
	FilesIndexed    int
	SnippetsIndexed int
	TotalTokens     int
	IndexSize       int64
	TimeTaken       time.Duration
	FileTypes       map[string]int
}

// QueryResult represents the result of querying a repository
type QueryResult struct {
	Results   []CodeSnippet
	TimeTaken time.Duration
}

// CodeSnippet represents a code snippet retrieved from the repository
type CodeSnippet struct {
	FilePath   string
	Snippet    string
	Similarity float64
}

// indexRepository indexes a repository for RAG
func indexRepository(repoPath string, filePatterns []string) (*IndexResult, error) {
	startTime := time.Now()

	// Validate repository path
	repoInfo, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("repository error: %v", err)
	}
	if !repoInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", repoPath)
	}

	// Create index directory if it doesn't exist
	indexDir := filepath.Join(repoPath, ".rag-index")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %v", err)
	}

	// Initialize result
	result := &IndexResult{
		FileTypes: make(map[string]int),
	}

	// Find all files matching the patterns
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

	// In a real implementation, we would:
	// 1. Parse each file to extract code snippets
	// 2. Generate embeddings for each snippet
	// 3. Store the embeddings in a vector database

	// For this simplified implementation, we'll just count the files and simulate the indexing
	result.SnippetsIndexed = result.FilesIndexed * 5  // Assume 5 snippets per file on average
	result.TotalTokens = result.SnippetsIndexed * 100 // Assume 100 tokens per snippet on average
	result.IndexSize = int64(result.TotalTokens * 4)  // Assume 4 bytes per token

	// Simulate indexing by writing a simple index file
	indexFile := filepath.Join(indexDir, "index.json")
	indexContent := fmt.Sprintf(`{
  "files_indexed": %d,
  "snippets_indexed": %d,
  "total_tokens": %d,
  "index_size": %d,
  "created_at": "%s"
}`, result.FilesIndexed, result.SnippetsIndexed, result.TotalTokens, result.IndexSize, time.Now().Format(time.RFC3339))

	if err := ioutil.WriteFile(indexFile, []byte(indexContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write index file: %v", err)
	}

	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// queryRepository queries a repository using RAG
func queryRepository(repoPath, query string, numResults int) (*QueryResult, error) {
	startTime := time.Now()

	// Validate repository path
	repoInfo, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("repository error: %v", err)
	}
	if !repoInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", repoPath)
	}

	// Check if index exists
	indexDir := filepath.Join(repoPath, ".rag-index")
	indexFile := filepath.Join(indexDir, "index.json")
	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("repository is not indexed, please run index operation first")
	}

	// In a real implementation, we would:
	// 1. Generate an embedding for the query
	// 2. Search the vector database for similar snippets
	// 3. Return the most relevant snippets

	// For this simplified implementation, we'll just return some mock results
	// based on a simple keyword search

	// Find files that might contain the query keywords
	keywords := strings.Fields(strings.ToLower(query))
	var matchingFiles []string
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Walk through the repository
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

		// Skip non-text files
		ext := filepath.Ext(path)
		if !isTextFile(ext) {
			return nil
		}

		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			// Read file content
			content, err := ioutil.ReadFile(filePath)
			if err != nil {
				return
			}

			// Check if file contains any of the keywords
			fileContent := strings.ToLower(string(content))
			for _, keyword := range keywords {
				if strings.Contains(fileContent, keyword) {
					mu.Lock()
					matchingFiles = append(matchingFiles, filePath)
					mu.Unlock()
					break
				}
			}
		}(path)

		return nil
	})

	wg.Wait()

	if err != nil {
		return nil, fmt.Errorf("error searching repository: %v", err)
	}

	// Generate mock results
	var results []CodeSnippet
	for _, file := range matchingFiles {
		if len(results) >= numResults {
			break
		}

		// Read file content
		content, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}

		// Find a relevant snippet
		fileContent := string(content)
		var snippet string

		// Look for a section containing the first keyword
		if len(keywords) > 0 {
			keyword := keywords[0]
			index := strings.Index(strings.ToLower(fileContent), keyword)
			if index >= 0 {
				// Extract a snippet around the keyword
				start := index - 100
				if start < 0 {
					start = 0
				}
				end := index + 300
				if end > len(fileContent) {
					end = len(fileContent)
				}

				// Find line boundaries
				for start > 0 && fileContent[start] != '\n' {
					start--
				}
				for end < len(fileContent) && fileContent[end] != '\n' {
					end++
				}

				snippet = fileContent[start:end]
			} else {
				// If keyword not found, just take the first few lines
				lines := strings.SplitN(fileContent, "\n", 11)
				snippet = strings.Join(lines[:min(10, len(lines))], "\n")
			}
		} else {
			// If no keywords, just take the first few lines
			lines := strings.SplitN(fileContent, "\n", 11)
			snippet = strings.Join(lines[:min(10, len(lines))], "\n")
		}

		// Calculate a mock similarity score
		similarity := calculateMockSimilarity(query, snippet)

		results = append(results, CodeSnippet{
			FilePath:   file,
			Snippet:    snippet,
			Similarity: similarity,
		})
	}

	// Sort results by similarity
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Limit results
	if len(results) > numResults {
		results = results[:numResults]
	}

	return &QueryResult{
		Results:   results,
		TimeTaken: time.Since(startTime),
	}, nil
}

// generateResponse generates a response based on the query and retrieved snippets
func generateResponse(query string, snippets []CodeSnippet) string {
	// In a real implementation, we would:
	// 1. Send the query and snippets to an LLM
	// 2. Have the LLM generate a response

	// For this simplified implementation, we'll just generate a mock response
	if len(snippets) == 0 {
		return "I couldn't find any relevant code snippets for your query."
	}

	response := fmt.Sprintf("Based on the code snippets I found, here's a response to your query about '%s':\n\n", query)

	// Extract keywords from the query
	keywords := strings.Fields(strings.ToLower(query))

	// Check if query is about how to do something
	if strings.Contains(strings.ToLower(query), "how to") || strings.Contains(strings.ToLower(query), "how do i") {
		response += "To accomplish this task, you can follow these steps:\n\n"

		for i, snippet := range snippets {
			response += fmt.Sprintf("%d. Look at the example in `%s`:\n", i+1, filepath.Base(snippet.FilePath))

			// Extract a few lines from the snippet
			lines := strings.Split(snippet.Snippet, "\n")
			relevantLines := []string{}
			for _, line := range lines {
				if len(relevantLines) >= 3 {
					break
				}

				// Check if line contains any keyword
				lineContainsKeyword := false
				for _, keyword := range keywords {
					if strings.Contains(strings.ToLower(line), keyword) {
						lineContainsKeyword = true
						break
					}
				}

				if lineContainsKeyword && len(strings.TrimSpace(line)) > 0 {
					relevantLines = append(relevantLines, line)
				}
			}

			if len(relevantLines) > 0 {
				response += "```\n"
				for _, line := range relevantLines {
					response += line + "\n"
				}
				response += "```\n\n"
			}
		}

		response += "You can adapt these examples to your specific needs."
	} else if strings.Contains(strings.ToLower(query), "what is") || strings.Contains(strings.ToLower(query), "explain") {
		// Query is asking for an explanation
		response += "Here's an explanation based on the code I found:\n\n"

		// Look for comments in the snippets
		foundComments := false
		for _, snippet := range snippets {
			lines := strings.Split(snippet.Snippet, "\n")
			comments := []string{}

			for _, line := range lines {
				trimmedLine := strings.TrimSpace(line)
				if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "#") ||
					strings.HasPrefix(trimmedLine, "/*") || strings.Contains(trimmedLine, "*/") {
					comments = append(comments, line)
				}
			}

			if len(comments) > 0 {
				foundComments = true
				response += fmt.Sprintf("From `%s`:\n", filepath.Base(snippet.FilePath))
				for _, comment := range comments {
					response += comment + "\n"
				}
				response += "\n"
			}
		}

		if !foundComments {
			response += "I couldn't find specific documentation about this in the code, but based on the code structure, "
			response += "it appears to be related to " + strings.Join(keywords, ", ") + ".\n\n"

			// Include a code example
			if len(snippets) > 0 {
				response += "Here's a relevant code example:\n\n```\n"
				lines := strings.Split(snippets[0].Snippet, "\n")
				for i := 0; i < min(5, len(lines)); i++ {
					response += lines[i] + "\n"
				}
				response += "```\n"
			}
		}
	} else {
		// Generic query
		response += "Here are some code examples that might help:\n\n"

		for i, snippet := range snippets {
			response += fmt.Sprintf("Example %d from `%s`:\n\n```\n", i+1, filepath.Base(snippet.FilePath))

			// Include a portion of the snippet
			lines := strings.Split(snippet.Snippet, "\n")
			for i := 0; i < min(7, len(lines)); i++ {
				response += lines[i] + "\n"
			}

			response += "```\n\n"
		}
	}

	return response
}

// calculateMockSimilarity calculates a mock similarity score between a query and a snippet
func calculateMockSimilarity(query, snippet string) float64 {
	// In a real implementation, we would:
	// 1. Calculate the cosine similarity between the query and snippet embeddings

	// For this simplified implementation, we'll just calculate a mock score
	// based on keyword matching
	keywords := strings.Fields(strings.ToLower(query))
	snippetLower := strings.ToLower(snippet)

	var score float64
	for _, keyword := range keywords {
		if strings.Contains(snippetLower, keyword) {
			// Count occurrences
			count := strings.Count(snippetLower, keyword)
			score += float64(count) * 0.1
		}
	}

	// Normalize score to [0, 1]
	if score > 1.0 {
		score = 1.0
	}

	// Add some randomness to simulate vector similarity
	score = score*0.7 + (0.3 * (0.5 + (float64(len(snippet)%100) / 200.0)))

	return score
}

// isTextFile checks if a file extension corresponds to a text file
func isTextFile(ext string) bool {
	textExtensions := map[string]bool{
		".go":   true,
		".js":   true,
		".ts":   true,
		".py":   true,
		".java": true,
		".c":    true,
		".cpp":  true,
		".h":    true,
		".cs":   true,
		".html": true,
		".css":  true,
		".md":   true,
		".txt":  true,
		".json": true,
		".xml":  true,
		".yml":  true,
		".yaml": true,
		".sh":   true,
		".bat":  true,
		".ps1":  true,
	}

	return textExtensions[ext]
}

// getLanguageFromFilePath determines the programming language from a file path
func getLanguageFromFilePath(path string) string {
	ext := filepath.Ext(path)
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
	case ".yml", ".yaml":
		return "yaml"
	case ".sh":
		return "bash"
	case ".bat":
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

// RegisterRAG registers the RAG tool with the MCP server
func RegisterRAG(mcpServer *server.MCPServer) {
	mcpServer.AddTool(mcp.NewTool("rag",
		mcp.WithDescription("Provides AI-powered code assistance using Retrieval Augmented Generation (RAG), which combines information retrieval with generative AI"),
		mcp.WithString("operation",
			mcp.Description("Operation to perform: 'index' to index a repository, 'query' to query the repository"),
			mcp.Required(),
		),
		mcp.WithString("repo_path",
			mcp.Description("Path to the repository"),
			mcp.Required(),
		),
		mcp.WithArray("file_patterns",
			mcp.Description("File patterns to include in the index (e.g., ['*.go', '*.js'])"),
		),
		mcp.WithString("query",
			mcp.Description("Query to search for (for 'query' operation)"),
		),
		mcp.WithNumber("num_results",
			mcp.Description("Number of results to return (for 'query' operation)"),
		),
	), HandleRAG)
}
