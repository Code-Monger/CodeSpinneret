package rag

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	workspace "github.com/Code-Monger/CodeSpinneret/pkg/workspace"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleRAG is the handler function for the RAG tool
func HandleRAG(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract session ID
	sessionID, _ := arguments["session_id"].(string)

	// Check if workspace is initialized
	_, exists := workspace.GetWorkspaceInfo(sessionID)
	if !exists {
		return nil, fmt.Errorf("workspace not initialized: please call the workspace tool with operation='initialize' first to create a session")
	}

	// Extract operation
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	switch operation {
	case "index":
		// Extract repository path
		repoPath, ok := arguments["repo_path"].(string)
		if !ok {
			return nil, fmt.Errorf("repo_path must be a string")
		}

		// Resolve repository path against workspace root directory
		workspaceInfo, _ := workspace.GetWorkspaceInfo(sessionID)
		repoPath = filepath.Join(workspaceInfo.RootDir, repoPath)
		log.Printf("[RAG] Using repository path: %s", repoPath)

		// Extract file patterns
		var filePatterns []string
		if filePatternArray, ok := arguments["file_patterns"].([]interface{}); ok {
			for _, pattern := range filePatternArray {
				if patternStr, ok := pattern.(string); ok {
					filePatterns = append(filePatterns, patternStr)
				}
			}
		}
		if len(filePatterns) == 0 {
			// Default patterns
			filePatterns = []string{"*.go", "*.js", "*.ts", "*.py", "*.java", "*.c", "*.cpp", "*.cs", "*.html", "*.css"}
		}

		// Index the repository
		indexResult, err := indexRepository(repoPath, filePatterns)
		if err != nil {
			return nil, fmt.Errorf("error indexing repository: %v", err)
		}

		// Format the result
		resultText := fmt.Sprintf("RAG Indexing Results\n\n")
		resultText += fmt.Sprintf("Repository: %s\n", repoPath)
		resultText += fmt.Sprintf("Files indexed: %d\n", indexResult.FilesIndexed)
		resultText += fmt.Sprintf("Code snippets: %d\n", indexResult.SnippetsIndexed)
		resultText += fmt.Sprintf("Total tokens: %d\n", indexResult.TotalTokens)
		resultText += fmt.Sprintf("Index size: %s\n", formatSize(indexResult.IndexSize))
		resultText += fmt.Sprintf("Time taken: %s\n\n", indexResult.TimeTaken)

		resultText += "File types:\n"
		for ext, count := range indexResult.FileTypes {
			resultText += fmt.Sprintf("- %s: %d files\n", ext, count)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	case "query":
		// Extract repository path
		repoPath, ok := arguments["repo_path"].(string)
		if !ok {
			return nil, fmt.Errorf("repo_path must be a string")
		}

		// Resolve repository path against workspace root directory
		workspaceInfo, _ := workspace.GetWorkspaceInfo(sessionID)
		repoPath = filepath.Join(workspaceInfo.RootDir, repoPath)
		log.Printf("[RAG] Using repository path: %s", repoPath)

		// Extract query
		query, ok := arguments["query"].(string)
		if !ok {
			return nil, fmt.Errorf("query must be a string")
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
			resultText += fmt.Sprintf("%d. File: %s (Similarity: %.2f)\n", i+1, result.FilePath, result.Similarity)
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

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// RegisterRAG registers the RAG tool with the MCP server
func RegisterRAG(mcpServer *server.MCPServer) {
	// Create the tool definition
	ragTool := mcp.NewTool("rag",
		mcp.WithDescription("Provides AI-powered code assistance using Retrieval Augmented Generation (RAG), which combines information retrieval with generative AI. Requires workspace initialization before use."),
		mcp.WithString("operation",
			mcp.Description("Operation to perform: 'index' to index a repository, 'query' to query the repository"),
			mcp.Required(),
		),
		mcp.WithString("repo_path",
			mcp.Description("Path to the repository"),
			mcp.Required(),
		),
		mcp.WithString("session_id",
			mcp.Description("Session ID to use for workspace initialization. Must be initialized with the workspace tool before using RAG."),
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
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("rag", HandleRAG)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(ragTool, wrappedHandler)

	// Log the registration
	log.Printf("[RAG] Registered rag tool")
}
