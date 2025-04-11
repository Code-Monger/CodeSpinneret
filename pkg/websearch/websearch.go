package websearch

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleWebSearch is the handler function for the web search tool
func HandleWebSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract query
	query, ok := arguments["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query must be a string")
	}

	// Extract number of results
	numResults := 10 // Default to 10 results
	if numResultsFloat, ok := arguments["num_results"].(float64); ok {
		numResults = int(numResultsFloat)
	}

	// Extract search engine
	engine, _ := arguments["engine"].(string)
	if engine == "" {
		engine = "duckduckgo" // Default to DuckDuckGo
	}

	// Extract safe search flag
	safeSearch := true // Default to safe search
	if safeSearchBool, ok := arguments["safe_search"].(bool); ok {
		safeSearch = safeSearchBool
	}

	// Perform the search
	var results *SearchResults
	var err error

	switch strings.ToLower(engine) {
	case "duckduckgo":
		results, err = searchDuckDuckGo(query, numResults, safeSearch)
	case "bing":
		results, err = searchBing(query, numResults, safeSearch)
	case "google":
		results, err = searchGoogle(query, numResults, safeSearch)
	default:
		return nil, fmt.Errorf("unsupported search engine: %s", engine)
	}

	if err != nil {
		return nil, fmt.Errorf("error performing search: %v", err)
	}

	// Format the results
	resultText := fmt.Sprintf("Search Results for '%s' using %s:\n\n", query, results.Engine)
	for i, result := range results.Results {
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Title)
		resultText += fmt.Sprintf("   URL: %s\n", result.URL)
		resultText += fmt.Sprintf("   Snippet: %s\n\n", result.Snippet)
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

// RegisterWebSearch registers the web search tool with the MCP server
func RegisterWebSearch(mcpServer *server.MCPServer) {
	// Create the tool definition
	webSearchTool := mcp.NewTool("websearch",
		mcp.WithDescription("Performs web searches to find information online, allowing AI models to access up-to-date information"),
		mcp.WithString("query",
			mcp.Description("The search query"),
			mcp.Required(),
		),
		mcp.WithNumber("num_results",
			mcp.Description("Number of results to return (default: 10)"),
		),
		mcp.WithString("engine",
			mcp.Description("Search engine to use (duckduckgo, bing, google) (default: duckduckgo)"),
		),
		mcp.WithBoolean("safe_search",
			mcp.Description("Whether to enable safe search (default: true)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("websearch", HandleWebSearch)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(webSearchTool, wrappedHandler)

	// Log the registration
	log.Printf("[WebSearch] Registered websearch tool")
}
