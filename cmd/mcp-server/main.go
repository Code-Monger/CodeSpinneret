package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/calculator"
	"github.com/Code-Monger/CodeSpinneret/pkg/cmdexec"
	"github.com/Code-Monger/CodeSpinneret/pkg/codeanalysis"
	"github.com/Code-Monger/CodeSpinneret/pkg/filesearch"
	"github.com/Code-Monger/CodeSpinneret/pkg/findcallers"
	"github.com/Code-Monger/CodeSpinneret/pkg/linecount"
	"github.com/Code-Monger/CodeSpinneret/pkg/patch"
	"github.com/Code-Monger/CodeSpinneret/pkg/rag"
	"github.com/Code-Monger/CodeSpinneret/pkg/screenshot"
	"github.com/Code-Monger/CodeSpinneret/pkg/searchreplace"
	"github.com/Code-Monger/CodeSpinneret/pkg/serverinfo"
	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/Code-Monger/CodeSpinneret/pkg/webfetch"
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
	dataDir      = flag.String("data-dir", filepath.Join(".", "data"), "Directory to store data files")
)

func main() {
	flag.Parse()

	// Create data directory if it doesn't exist
	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

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

	// Initialize stats service
	if err := stats.InitStatsManager(*dataDir); err != nil {
		log.Fatalf("Failed to initialize stats manager: %v", err)
	}

	// Register tools and resources
	calculator.RegisterCalculator(mcpServer)
	serverinfo.RegisterServerInfo(mcpServer)
	filesearch.RegisterFileSearch(mcpServer)
	cmdexec.RegisterCommandExecution(mcpServer)
	searchreplace.RegisterSearchReplace(mcpServer)
	screenshot.RegisterScreenshot(mcpServer)
	websearch.RegisterWebSearch(mcpServer)
	webfetch.Register(mcpServer)
	rag.RegisterRAG(mcpServer)
	codeanalysis.RegisterCodeAnalysis(mcpServer)
	patch.RegisterPatch(mcpServer)
	linecount.RegisterLineCount(mcpServer)
	findcallers.RegisterFindCallers(mcpServer)

	// Register stats tool
	if err := stats.RegisterStats(mcpServer, *dataDir); err != nil {
		log.Fatalf("Failed to register stats tool: %v", err)
	}

	// Create the SSE server
	baseURLValue := *baseURL
	if baseURLValue == "" {
		baseURLValue = fmt.Sprintf("http://localhost:%d", *port)
	}

	// Create SSE server
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
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Printf("[Server] Starting MCP server on port %d...", *port)
		log.Printf("[Server] Base URL: %s", baseURLValue)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-stop

	// Create a deadline for shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown the server
	log.Println("[Server] Shutting down server...")

	// Print final stats before shutdown
	if statsManager := stats.GetStatsManager(); statsManager != nil {
		sessionStats := statsManager.GetSessionStats()
		persistentStats := statsManager.GetPersistentStats()
		statsText := stats.FormatStats(sessionStats, persistentStats)
		log.Printf("[Server] Final server statistics:\n%s", statsText)
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[Server] Server shutdown failed: %v", err)
	}
	log.Println("[Server] Server stopped")
}
