package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/calculator"
	"github.com/Code-Monger/CodeSpinneret/pkg/cmdexec"
	"github.com/Code-Monger/CodeSpinneret/pkg/filesearch"
	"github.com/Code-Monger/CodeSpinneret/pkg/rag"
	"github.com/Code-Monger/CodeSpinneret/pkg/screenshot"
	"github.com/Code-Monger/CodeSpinneret/pkg/searchreplace"
	"github.com/Code-Monger/CodeSpinneret/pkg/serverinfo"
	"github.com/Code-Monger/CodeSpinneret/pkg/websearch"
	"github.com/mark3labs/mcp-go/server"
)

var (
	port         = flag.Int("port", 8080, "Port to listen on")
	baseURL      = flag.String("baseurl", "", "Base URL for the server (e.g., http://localhost:8080)")
	serverName   = flag.String("name", "CodeSpinneret MCP Server", "Server name")
	serverVer    = flag.String("version", "1.0.0", "Server version")
	timeoutSecs  = flag.Int("timeout", 300, "Server timeout in seconds")
	instructions = flag.String("instructions", "This is a Model Context Protocol server implementation.", "Server instructions")
)

func main() {
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutSecs)*time.Second)
	defer cancel()

	// Create the MCP server
	mcpServer := server.NewMCPServer(
		*serverName,
		*serverVer,
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithInstructions(*instructions),
	)

	// Register tools and resources
	calculator.RegisterCalculator(mcpServer)
	serverinfo.RegisterServerInfo(mcpServer)
	filesearch.RegisterFileSearch(mcpServer)
	cmdexec.RegisterCommandExecution(mcpServer)
	searchreplace.RegisterSearchReplace(mcpServer)
	screenshot.RegisterScreenshot(mcpServer)
	websearch.RegisterWebSearch(mcpServer)
	rag.RegisterRAG(mcpServer)

	// Create the SSE server
	baseURLValue := *baseURL
	if baseURLValue == "" {
		baseURLValue = fmt.Sprintf("http://localhost:%d", *port)
	}

	sseServer := server.NewSSEServer(
		mcpServer,
		server.WithBaseURL(baseURLValue),
		server.WithSSEEndpoint("/"),
		server.WithMessageEndpoint("/messages"),
	)

	// Set up HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: sseServer,
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting MCP server on port %d...", *port)
		log.Printf("Base URL: %s", baseURLValue)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for termination signal or context timeout
	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	case <-ctx.Done():
		log.Printf("Server timeout reached")
	}

	// Create a shutdown context with a timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Perform graceful shutdown
	log.Println("Shutting down server...")
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server shutdown complete")
}
