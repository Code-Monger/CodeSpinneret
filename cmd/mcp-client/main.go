package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	serverURL   = flag.String("server", "http://localhost:8080", "MCP server URL")
	timeoutSecs = flag.Int("timeout", 60, "Client timeout in seconds")
	testTool    = flag.String("tool", "calculator", "Tool to test (calculator, filesearch, cmdexec, searchreplace)")
)

func main() {
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSecs)*time.Second)
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a context cancellation for signal handling
	signalCtx, signalCancel := context.WithCancel(context.Background())
	defer signalCancel()

	// Handle termination signals
	go func() {
		select {
		case sig := <-sigChan:
			log.Printf("Received signal: %v", sig)
			signalCancel()
		case <-ctx.Done():
			log.Printf("Client timeout reached")
		}
	}()

	// Create the SSE client
	log.Printf("Connecting to MCP server at %s...", *serverURL)
	sseClient, err := client.NewSSEMCPClient(*serverURL)
	if err != nil {
		log.Fatalf("Failed to create SSE client: %v", err)
	}

	// Start the client
	if err := sseClient.Start(signalCtx); err != nil {
		log.Fatalf("Failed to start SSE client: %v", err)
	}

	// Initialize the client
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION

	initResult, err := sseClient.Initialize(signalCtx, initReq)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	log.Printf("Connected to server successfully")
	log.Printf("Server capabilities: %+v", initResult.Capabilities)

	// List available resources
	resourcesResult, err := sseClient.ListResources(signalCtx, mcp.ListResourcesRequest{})
	if err != nil {
		log.Printf("Failed to list resources: %v", err)
	} else {
		log.Printf("Available resources (%d):", len(resourcesResult.Resources))
		for _, resource := range resourcesResult.Resources {
			log.Printf("  - %s (%s)", resource.Name, resource.URI)
		}
	}

	// List available tools
	toolsResult, err := sseClient.ListTools(signalCtx, mcp.ListToolsRequest{})
	if err != nil {
		log.Printf("Failed to list tools: %v", err)
	} else {
		log.Printf("Available tools (%d):", len(toolsResult.Tools))
		for _, tool := range toolsResult.Tools {
			log.Printf("  - %s: %s", tool.Name, tool.Description)
		}
	}

	// Test the specified tool if available
	switch *testTool {
	case "calculator":
		found := false
		for _, tool := range toolsResult.Tools {
			if tool.Name == "calculator" {
				found = true
				break
			}
		}

		if found {
			log.Println("Testing calculator tool...")
			testCalculator(signalCtx, sseClient)
		} else {
			log.Println("Calculator tool not found on server")
		}
	case "filesearch":
		found := false
		for _, tool := range toolsResult.Tools {
			if tool.Name == "filesearch" {
				found = true
				break
			}
		}

		if found {
			log.Println("Testing file search tool...")
			testFileSearch(signalCtx, sseClient)
		} else {
			log.Println("File search tool not found on server")
		}
	case "cmdexec":
		found := false
		for _, tool := range toolsResult.Tools {
			if tool.Name == "cmdexec" {
				found = true
				break
			}
		}

		if found {
			log.Println("Testing command execution tool...")
			testCommandExecution(signalCtx, sseClient)
		} else {
			log.Println("Command execution tool not found on server")
		}
	case "searchreplace":
		found := false
		for _, tool := range toolsResult.Tools {
			if tool.Name == "searchreplace" {
				found = true
				break
			}
		}

		if found {
			log.Println("Testing search replace tool...")
			testSearchReplace(signalCtx, sseClient)
		} else {
			log.Println("Search replace tool not found on server")
		}
	}

	// Read server info resource if available
	found := false
	for _, resource := range resourcesResult.Resources {
		if resource.URI == "server://info" {
			found = true
			break
		}
	}

	if found {
		log.Println("Reading server info resource...")
		readServerInfo(signalCtx, sseClient)
	} else {
		log.Println("Server info resource not found on server")
	}

	log.Println("Client operations completed successfully")
}

// testCalculator tests the calculator tool with various operations
func testCalculator(ctx context.Context, c client.MCPClient) {
	operations := []struct {
		op string
		a  float64
		b  float64
	}{
		{"add", 5, 3},
		{"subtract", 10, 4},
		{"multiply", 6, 7},
		{"divide", 20, 5},
	}

	for _, op := range operations {
		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "calculator"
		callReq.Params.Arguments = map[string]interface{}{
			"operation": op.op,
			"a":         op.a,
			"b":         op.b,
		}

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call calculator with %s: %v", op.op, err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Calculator %s result: %s", op.op, textContent.Text)
			}
		}
	}
}

// readServerInfo reads the server info resource
func readServerInfo(ctx context.Context, c client.MCPClient) {
	readReq := mcp.ReadResourceRequest{}
	readReq.Params.URI = "server://info"

	result, err := c.ReadResource(ctx, readReq)
	if err != nil {
		log.Printf("Failed to read server info: %v", err)
		return
	}

	if len(result.Contents) > 0 {
		if textContent, ok := result.Contents[0].(mcp.TextResourceContents); ok {
			log.Printf("Server Info:\n%s", textContent.Text)
		}
	}
}

// testFileSearch tests the file search tool with various search criteria
func testFileSearch(ctx context.Context, c client.MCPClient) {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Basic directory listing",
			arguments: map[string]interface{}{
				"directory": ".",
				"pattern":   "*.go",
				"recursive": false,
			},
		},
		{
			name: "Recursive search",
			arguments: map[string]interface{}{
				"directory": ".",
				"pattern":   "*.go",
				"recursive": true,
			},
		},
		{
			name: "Content search",
			arguments: map[string]interface{}{
				"directory":       ".",
				"pattern":         "*.go",
				"recursive":       true,
				"content_pattern": "func.*\\(",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running file search test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "filesearch"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call filesearch: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("File search result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}
}

// testCommandExecution tests the command execution tool with various commands
func testCommandExecution(ctx context.Context, c client.MCPClient) {
	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Simple echo command",
			arguments: map[string]interface{}{
				"command": "echo Hello, World!",
				"timeout": 5.0,
			},
		},
		{
			name: "Directory listing",
			arguments: map[string]interface{}{
				"command": "dir",
				"timeout": 5.0,
			},
		},
		{
			name: "Current working directory",
			arguments: map[string]interface{}{
				"command": "cd",
				"timeout": 5.0,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running command execution test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "cmdexec"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call cmdexec: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Command execution result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}
}

// testSearchReplace tests the search replace tool with various operations
func testSearchReplace(ctx context.Context, c client.MCPClient) {
	// Create a temporary test file
	tempDir := os.TempDir()
	testFilePath := filepath.Join(tempDir, "mcp_test_search_replace.txt")

	testContent := `This is a test file for search and replace.
It contains multiple lines of text.
We will search for specific patterns and replace them.
This line has the word 'test' in it twice for testing.
The end of the test file.`

	err := ioutil.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		log.Printf("Failed to create test file: %v", err)
		return
	}

	defer func() {
		// Clean up the test file
		os.Remove(testFilePath)
		log.Println("Test file removed")
	}()

	log.Printf("Created test file at: %s", testFilePath)

	// Define test cases
	testCases := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name: "Simple string replacement (preview)",
			arguments: map[string]interface{}{
				"directory":      filepath.Dir(testFilePath),
				"file_pattern":   filepath.Base(testFilePath),
				"search_pattern": "test",
				"replacement":    "EXAMPLE",
				"use_regex":      false,
				"recursive":      false,
				"preview":        true,
				"case_sensitive": true,
			},
		},
		{
			name: "Regex replacement (preview)",
			arguments: map[string]interface{}{
				"directory":      filepath.Dir(testFilePath),
				"file_pattern":   filepath.Base(testFilePath),
				"search_pattern": "t[a-z]{3}",
				"replacement":    "MATCH",
				"use_regex":      true,
				"recursive":      false,
				"preview":        true,
				"case_sensitive": false,
			},
		},
		{
			name: "Actual replacement",
			arguments: map[string]interface{}{
				"directory":      filepath.Dir(testFilePath),
				"file_pattern":   filepath.Base(testFilePath),
				"search_pattern": "line",
				"replacement":    "ROW",
				"use_regex":      false,
				"recursive":      false,
				"preview":        false,
				"case_sensitive": true,
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		log.Printf("Running search replace test: %s", tc.name)

		callReq := mcp.CallToolRequest{}
		callReq.Params.Name = "searchreplace"
		callReq.Params.Arguments = tc.arguments

		result, err := c.CallTool(ctx, callReq)
		if err != nil {
			log.Printf("Failed to call searchreplace: %v", err)
			continue
		}

		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(mcp.TextContent); ok {
				log.Printf("Search replace result:\n%s", textContent.Text)
			}
		}

		// Add a small delay between tests
		time.Sleep(500 * time.Millisecond)
	}

	// Read the modified file to show changes
	modifiedContent, err := ioutil.ReadFile(testFilePath)
	if err != nil {
		log.Printf("Failed to read modified file: %v", err)
		return
	}

	log.Printf("Final file content after modifications:\n%s", string(modifiedContent))
}
