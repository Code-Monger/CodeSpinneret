package websearch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

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
	numResults := 10 // Default
	if numResultsFloat, ok := arguments["num_results"].(float64); ok {
		numResults = int(numResultsFloat)
	}

	// Extract search engine
	engine, _ := arguments["engine"].(string)
	if engine == "" {
		engine = "duckduckgo" // Default to DuckDuckGo
	}

	// Extract safe search
	safeSearch := true // Default to safe search
	if safeSearchBool, ok := arguments["safe_search"].(bool); ok {
		safeSearch = safeSearchBool
	}

	// Perform the search
	results, err := performSearch(query, numResults, engine, safeSearch)
	if err != nil {
		return nil, fmt.Errorf("error performing web search: %v", err)
	}

	// Format the results
	resultText := fmt.Sprintf("Web Search Results for: %s\n\n", query)
	resultText += fmt.Sprintf("Search Engine: %s\n", results.Engine)
	resultText += fmt.Sprintf("Results: %d\n\n", len(results.Results))

	for i, result := range results.Results {
		resultText += fmt.Sprintf("%d. %s\n", i+1, result.Title)
		resultText += fmt.Sprintf("   URL: %s\n", result.URL)
		if result.Description != "" {
			resultText += fmt.Sprintf("   Description: %s\n", result.Description)
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

// SearchResult represents a single search result
type SearchResult struct {
	Title       string
	URL         string
	Description string
}

// SearchResults represents the results of a web search
type SearchResults struct {
	Query   string
	Engine  string
	Results []SearchResult
}

// performSearch performs a web search using the specified engine
func performSearch(query string, numResults int, engine string, safeSearch bool) (*SearchResults, error) {
	switch strings.ToLower(engine) {
	case "duckduckgo":
		return searchDuckDuckGo(query, numResults, safeSearch)
	case "bing":
		return searchBing(query, numResults, safeSearch)
	case "google":
		return searchGoogle(query, numResults, safeSearch)
	default:
		return nil, fmt.Errorf("unsupported search engine: %s", engine)
	}
}

// searchDuckDuckGo performs a search using DuckDuckGo
func searchDuckDuckGo(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	// DuckDuckGo doesn't have an official API, so we'll use their lite version
	baseURL := "https://lite.duckduckgo.com/lite"

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create form data
	formData := url.Values{}
	formData.Set("q", query)
	if !safeSearch {
		formData.Set("kp", "-2") // Disable safe search
	}

	// Create request
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse HTML response (simplified)
	// In a real implementation, you would use a proper HTML parser
	results := parseSimplifiedDuckDuckGoResults(string(body), numResults)

	return &SearchResults{
		Query:   query,
		Engine:  "DuckDuckGo",
		Results: results,
	}, nil
}

// parseSimplifiedDuckDuckGoResults parses the HTML response from DuckDuckGo
// This is a simplified implementation and would need a proper HTML parser in production
func parseSimplifiedDuckDuckGoResults(html string, numResults int) []SearchResult {
	// This is a simplified implementation
	// In a real implementation, you would use a proper HTML parser
	// For demonstration purposes, we'll return some mock results
	mockResults := []SearchResult{
		{
			Title:       "Example Search Result 1",
			URL:         "https://example.com/result1",
			Description: "This is an example search result description.",
		},
		{
			Title:       "Example Search Result 2",
			URL:         "https://example.com/result2",
			Description: "Another example search result description.",
		},
		{
			Title:       "Example Search Result 3",
			URL:         "https://example.com/result3",
			Description: "Yet another example search result description.",
		},
	}

	// Limit the number of results
	if numResults < len(mockResults) {
		return mockResults[:numResults]
	}
	return mockResults
}

// searchBing performs a search using Bing
func searchBing(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	// Note: In a real implementation, you would use the Bing Search API
	// which requires an API key from Microsoft Azure
	// For demonstration purposes, we'll return some mock results
	mockResults := []SearchResult{
		{
			Title:       "Bing Search Result 1",
			URL:         "https://example.com/bing1",
			Description: "This is a Bing search result description.",
		},
		{
			Title:       "Bing Search Result 2",
			URL:         "https://example.com/bing2",
			Description: "Another Bing search result description.",
		},
	}

	// Limit the number of results
	if numResults < len(mockResults) {
		mockResults = mockResults[:numResults]
	}

	return &SearchResults{
		Query:   query,
		Engine:  "Bing",
		Results: mockResults,
	}, nil
}

// searchGoogle performs a search using Google
func searchGoogle(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	// Note: In a real implementation, you would use the Google Custom Search API
	// which requires an API key and a Custom Search Engine ID
	// For demonstration purposes, we'll return some mock results
	mockResults := []SearchResult{
		{
			Title:       "Google Search Result 1",
			URL:         "https://example.com/google1",
			Description: "This is a Google search result description.",
		},
		{
			Title:       "Google Search Result 2",
			URL:         "https://example.com/google2",
			Description: "Another Google search result description.",
		},
		{
			Title:       "Google Search Result 3",
			URL:         "https://example.com/google3",
			Description: "Yet another Google search result description.",
		},
	}

	// Limit the number of results
	if numResults < len(mockResults) {
		mockResults = mockResults[:numResults]
	}

	return &SearchResults{
		Query:   query,
		Engine:  "Google",
		Results: mockResults,
	}, nil
}

// RegisterWebSearch registers the web search tool with the MCP server
func RegisterWebSearch(mcpServer *server.MCPServer) {
	mcpServer.AddTool(mcp.NewTool("websearch",
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
	), HandleWebSearch)
}
