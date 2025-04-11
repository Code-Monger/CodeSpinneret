package webfetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Handler handles web fetch requests
func Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract URL
	urlStr, ok := arguments["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url must be a string")
	}

	// Extract include_images
	includeImages := false
	if includeImagesVal, ok := arguments["include_images"].(bool); ok {
		includeImages = includeImagesVal
	}

	// Extract timeout
	timeout := 0
	switch v := arguments["timeout"].(type) {
	case float64:
		timeout = int(v)
	case int:
		timeout = v
	}

	// Create request
	webFetchRequest := WebFetchRequest{
		URL:           urlStr,
		IncludeImages: includeImages,
		Timeout:       timeout,
	}

	// Fetch the web page
	response, err := FetchWebPage(webFetchRequest)
	if err != nil {
		return nil, err
	}

	// Create a formatted response
	var resultText string
	if response.Error != "" {
		resultText = fmt.Sprintf("Error fetching %s: %s", response.URL, response.Error)
	} else {
		resultText = fmt.Sprintf("Successfully fetched %s\n\nStatus Code: %d\nContent Type: %s\n\nContent:\n%s",
			response.URL, response.StatusCode, response.ContentType, truncateContent(response.Content, 5000))
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

// truncateContent truncates content to a maximum length
func truncateContent(content string, maxLength int) string {
	if len(content) <= maxLength {
		return content
	}
	return content[:maxLength] + "... [truncated]"
}

// FetchWebPage fetches a web page and returns its content
func FetchWebPage(request WebFetchRequest) (*WebFetchResponse, error) {
	config := GetConfig()

	// Validate URL
	if request.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	// Add scheme if missing
	if !strings.HasPrefix(request.URL, "http://") && !strings.HasPrefix(request.URL, "https://") {
		request.URL = "https://" + request.URL
	}

	// Set timeout
	timeout := config.DefaultTimeout
	if request.Timeout > 0 {
		timeout = request.Timeout
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", request.URL, nil)
	if err != nil {
		return &WebFetchResponse{
			URL:   request.URL,
			Error: fmt.Sprintf("failed to create request: %v", err),
		}, nil
	}

	// Set headers
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Cache-Control", "max-age=0")

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return &WebFetchResponse{
			URL:   request.URL,
			Error: fmt.Sprintf("failed to fetch URL: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	// Create response
	response := &WebFetchResponse{
		URL:         request.URL,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Headers:     make(map[string]string),
	}

	// Add headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		response.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		return response, nil
	}

	// Read response body with size limit
	limitedReader := io.LimitReader(resp.Body, int64(config.MaxContentSize))
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		response.Error = fmt.Sprintf("failed to read response body: %v", err)
		return response, nil
	}

	// Set content
	response.Content = string(body)

	// Remove images if not requested
	if !request.IncludeImages && strings.Contains(response.ContentType, "text/html") {
		response.Content = removeImages(response.Content)
	}

	return response, nil
}

// removeImages removes image tags from HTML content
func removeImages(content string) string {
	// Simple regex-like replacement for <img> tags
	result := content

	// Replace <img> tags
	for {
		imgStart := strings.Index(result, "<img")
		if imgStart == -1 {
			break
		}

		imgEnd := strings.Index(result[imgStart:], ">")
		if imgEnd == -1 {
			break
		}

		imgEnd += imgStart + 1
		result = result[:imgStart] + result[imgEnd:]
	}

	return result
}

// Register registers the webfetch handler with the MCP server
func Register(mcpServer *server.MCPServer) {
	// Create the tool definition
	webfetchTool := mcp.NewTool("webfetch",
		mcp.WithDescription("Fetches the content of a web page given a URL"),
		mcp.WithString("url",
			mcp.Description("The URL to fetch"),
			mcp.Required(),
		),
		mcp.WithBoolean("include_images",
			mcp.Description("Whether to include images in the response"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Timeout in seconds"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("webfetch", Handler)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(webfetchTool, wrappedHandler)
}
