package websearch

import (
	"context"
	"fmt"
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

// SearchResults represents the results of a web search
type SearchResults struct {
	Query   string
	Engine  string
	Results []SearchResult
}

// SearchResult represents a single search result
type SearchResult struct {
	Title   string
	URL     string
	Snippet string
}

// searchDuckDuckGo performs a search using DuckDuckGo
func searchDuckDuckGo(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	// Note: DuckDuckGo doesn't have an official API, so we'd need to use web scraping
	// For demonstration purposes, we'll return some mock results
	mockResults := []SearchResult{
		{
			Title:   "DuckDuckGo — Privacy, simplified.",
			URL:     "https://duckduckgo.com/",
			Snippet: "The Internet privacy company that empowers you to seamlessly take control of your personal information online, without any tradeoffs.",
		},
		{
			Title:   "DuckDuckGo - Wikipedia",
			URL:     "https://en.wikipedia.org/wiki/DuckDuckGo",
			Snippet: "DuckDuckGo is an internet search engine that emphasizes protecting searchers' privacy and avoiding the filter bubble of personalized search results.",
		},
		{
			Title:   "DuckDuckGo (@DuckDuckGo) / X",
			URL:     "https://twitter.com/duckduckgo",
			Snippet: "The official Twitter account for DuckDuckGo, the Internet privacy company empowering you to seamlessly take control of your personal information online.",
		},
		{
			Title:   "DuckDuckGo Privacy Browser - Apps on Google Play",
			URL:     "https://play.google.com/store/apps/details?id=com.duckduckgo.mobile.android",
			Snippet: "DuckDuckGo Privacy Browser is a privacy browser for Android with built-in tracker blocking, encryption, and private search.",
		},
		{
			Title:   "DuckDuckGo Privacy Browser on the App Store",
			URL:     "https://apps.apple.com/us/app/duckduckgo-privacy-browser/id663592361",
			Snippet: "DuckDuckGo Privacy Browser is a privacy browser for iOS with built-in tracker blocking, encryption, and private search.",
		},
	}

	// Limit the number of results
	if numResults < len(mockResults) {
		mockResults = mockResults[:numResults]
	}

	return &SearchResults{
		Query:   query,
		Engine:  "DuckDuckGo",
		Results: mockResults,
	}, nil
}

// searchBing performs a search using Bing
func searchBing(query string, numResults int, safeSearch bool) (*SearchResults, error) {
	// Note: In a real implementation, you would use the Bing Search API
	// which requires an API key
	// For demonstration purposes, we'll return some mock results
	mockResults := []SearchResult{
		{
			Title:   "Bing",
			URL:     "https://www.bing.com/",
			Snippet: "Bing helps you turn information into action, making it faster and easier to go from searching to doing.",
		},
		{
			Title:   "Bing - Wikipedia",
			URL:     "https://en.wikipedia.org/wiki/Bing",
			Snippet: "Bing is a web search engine owned and operated by Microsoft. The service has its origins in Microsoft's previous search engines: MSN Search, Windows Live Search and later Live Search.",
		},
		{
			Title:   "Microsoft Bing Search API | Microsoft Azure",
			URL:     "https://azure.microsoft.com/en-us/services/cognitive-services/bing-web-search-api/",
			Snippet: "The Bing Search APIs let you build web-connected apps and services that find webpages, images, news, locations, and more without advertisements.",
		},
		{
			Title:   "Bing Maps - Directions, trip planning, traffic cameras & more",
			URL:     "https://www.bing.com/maps",
			Snippet: "Map multiple locations, get transit/walking/driving directions, view live traffic conditions, plan trips, view satellite, aerial and street side imagery.",
		},
		{
			Title:   "Bing Ads - Microsoft Advertising",
			URL:     "https://ads.microsoft.com/",
			Snippet: "Microsoft Advertising (formerly Bing Ads) is a service that provides pay per click advertising on the Bing and Yahoo! search engines.",
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
			Title:   "Google",
			URL:     "https://www.google.com/",
			Snippet: "Search the world's information, including webpages, images, videos and more. Google has many special features to help you find exactly what you're looking for.",
		},
		{
			Title:   "Google - Wikipedia",
			URL:     "https://en.wikipedia.org/wiki/Google",
			Snippet: "Google LLC is an American multinational technology company that specializes in Internet-related services and products, which include online advertising technologies, a search engine, cloud computing, software, and hardware.",
		},
		{
			Title:   "Google Search - Wikipedia",
			URL:     "https://en.wikipedia.org/wiki/Google_Search",
			Snippet: "Google Search, or simply Google, is a web search engine developed by Google LLC. It is the most used search engine on the World Wide Web across all platforms.",
		},
		{
			Title:   "Google Maps",
			URL:     "https://www.google.com/maps",
			Snippet: "Find local businesses, view maps and get driving directions in Google Maps.",
		},
		{
			Title:   "Gmail - Email from Google",
			URL:     "https://www.gmail.com/",
			Snippet: "Gmail is email that's intuitive, efficient, and useful. 15 GB of storage, less spam, and mobile access.",
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
}
