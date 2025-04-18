package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Code-Monger/CodeSpinneret/cmd/mcp-client/tools"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// Client represents the MCP client application
type Client struct {
	serverURL string
	mcpClient client.MCPClient
}

// NewClient creates a new MCP client
func NewClient(serverURL string) *Client {
	return &Client{
		serverURL: serverURL,
	}
}

// Run initializes and runs the client with the specified tool test
func (c *Client) Run(ctx context.Context, testTool string) error {
	// Create the SSE client
	log.Printf("Connecting to MCP server at %s...", c.serverURL)
	sseClient, err := client.NewSSEMCPClient(c.serverURL)
	if err != nil {
		return fmt.Errorf("failed to create SSE client: %v", err)
	}

	// Start the client
	if err := sseClient.Start(ctx); err != nil {
		return fmt.Errorf("failed to start SSE client: %v", err)
	}

	// Store the client
	c.mcpClient = sseClient

	// Initialize the client
	if err := c.initialize(ctx); err != nil {
		return err
	}

	// List available resources and tools
	resourcesResult, toolsResult, err := c.listResourcesAndTools(ctx)
	if err != nil {
		return err
	}

	// Test the specified tool if available
	if err := c.testTool(ctx, testTool, toolsResult); err != nil {
		return err
	}

	// Read server info resource if available
	c.readServerInfoIfAvailable(ctx, resourcesResult)

	return nil
}

// initialize initializes the MCP client
func (c *Client) initialize(ctx context.Context) error {
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION

	initResult, err := c.mcpClient.Initialize(ctx, initReq)
	if err != nil {
		return fmt.Errorf("failed to initialize client: %v", err)
	}

	log.Printf("Connected to server successfully")
	log.Printf("Server capabilities: %+v", initResult.Capabilities)
	return nil
}

// listResourcesAndTools lists available resources and tools
func (c *Client) listResourcesAndTools(ctx context.Context) (*mcp.ListResourcesResult, *mcp.ListToolsResult, error) {
	// List available resources
	resourcesResult, err := c.mcpClient.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list resources: %v", err)
	}

	log.Printf("Available resources (%d):", len(resourcesResult.Resources))
	for _, resource := range resourcesResult.Resources {
		log.Printf("  - %s (%s)", resource.Name, resource.URI)
	}

	// List available tools
	toolsResult, err := c.mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list tools: %v", err)
	}

	log.Printf("Available tools (%d):", len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		log.Printf("  - %s: %s", tool.Name, tool.Description)
	}

	return resourcesResult, toolsResult, nil
}

// testTool tests the specified tool
func (c *Client) testTool(ctx context.Context, testTool string, toolsResult *mcp.ListToolsResult) error {
	// Special case for "all" - test all available tools
	if testTool == "all" {
		return c.testAllTools(ctx, toolsResult)
	}

	// Check if the tool is available
	found := false
	for _, tool := range toolsResult.Tools {
		if tool.Name == testTool {
			found = true
			break
		}
	}

	if !found {
		log.Printf("%s tool not found on server", testTool)
		return nil
	}

	// Test the tool
	log.Printf("Testing %s tool...", testTool)

	switch testTool {
	case "calculator":
		return tools.TestCalculator(ctx, c.mcpClient)
	case "filesearch":
		return tools.TestFileSearch(ctx, c.mcpClient)
	case "cmdexec":
		return tools.TestCommandExecution(ctx, c.mcpClient)
	case "shell":
		return tools.TestShell(ctx, c.mcpClient)
	case "searchreplace":
		return tools.TestSearchReplace(ctx, c.mcpClient)
	case "screenshot":
		return tools.TestScreenshot(ctx, c.mcpClient)
	case "websearch":
		return tools.TestWebSearch(ctx, c.mcpClient)
	case "webfetch":
		return tools.TestWebFetch(ctx, c.mcpClient)
	case "rag":
		return tools.TestRAG(ctx, c.mcpClient)
	case "codeanalysis":
		return tools.TestCodeAnalysis(ctx, c.mcpClient)
	case "patch":
		return tools.TestPatch(ctx, c.mcpClient)
	case "linecount":
		return tools.TestLineCount(ctx, c.mcpClient)
	case "findcallers":
		return tools.TestFindCallers(ctx, c.mcpClient)
	case "findfunc":
		return tools.TestFindFunc(ctx, c.mcpClient)
	case "funcdef":
		return tools.TestFuncDef(ctx, c.mcpClient)
	case "spellcheck":
		return tools.TestSpellCheck(ctx, c.mcpClient)
	case "stats":
		return tools.TestStats(ctx, c.mcpClient)
	case "workspace":
		return tools.TestWorkspace(ctx, c.mcpClient)
	case "workspace_integration":
		return tools.TestWorkspaceIntegration(ctx, c.mcpClient)
	default:
		return fmt.Errorf("unknown tool: %s", testTool)
	}
}

// testAllTools tests all available tools
func (c *Client) testAllTools(ctx context.Context, toolsResult *mcp.ListToolsResult) error {
	log.Printf("Testing all available tools...")

	// Define the order of tools to test
	allTools := []string{
		"workspace", // Test workspace first as other tools may depend on it
		"calculator",
		"filesearch",
		"cmdexec",
		"shell",
		"searchreplace",
		"screenshot",
		"websearch",
		"webfetch",
		"rag",
		"codeanalysis",
		"patch",
		"linecount",
		"findcallers",
		"findfunc",
		"funcdef",
		"spellcheck",
		"stats",
		"workspace_integration", // Test integration between workspace and other tools
	}

	// Track overall statistics
	startTime := time.Now()
	var successCount, failureCount int

	// Test each tool
	for _, toolName := range allTools {
		// Check if the tool is available
		available := false
		for _, tool := range toolsResult.Tools {
			if tool.Name == toolName {
				available = true
				break
			}
		}

		if !available {
			log.Printf("Skipping %s (not available on server)", toolName)
			continue
		}

		// Create a separator for better readability
		log.Printf("\n%s\n=== TESTING %s ===\n%s\n",
			"------------------------------------------------",
			toolName,
			"------------------------------------------------")

		// Test the tool
		toolStartTime := time.Now()
		err := c.testTool(ctx, toolName, toolsResult)
		toolDuration := time.Since(toolStartTime)

		if err != nil {
			log.Printf("❌ %s test FAILED in %v: %v", toolName, toolDuration, err)
			failureCount++
		} else {
			log.Printf("✅ %s test PASSED in %v", toolName, toolDuration)
			successCount++
		}
	}

	// Print summary
	totalDuration := time.Since(startTime)
	log.Printf("\n%s\n=== TEST SUMMARY ===\n%s",
		"================================================",
		"================================================")
	log.Printf("Total tools tested: %d", successCount+failureCount)
	log.Printf("Successful tests: %d", successCount)
	log.Printf("Failed tests: %d", failureCount)
	log.Printf("Total duration: %v", totalDuration)
	log.Printf("================================================")

	return nil
}

// readServerInfoIfAvailable reads the server info resource if available
func (c *Client) readServerInfoIfAvailable(ctx context.Context, resourcesResult *mcp.ListResourcesResult) {
	// Check if the server info resource is available
	found := false
	for _, resource := range resourcesResult.Resources {
		if resource.URI == "server://info" {
			found = true
			break
		}
	}

	if !found {
		log.Println("Server info resource not found on server")
		return
	}

	// Read server info
	log.Println("Reading server info resource...")
	ReadServerInfo(ctx, c.mcpClient)
}
