package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	serverURL   = flag.String("server", "http://localhost:8080", "MCP server URL")
	timeoutSecs = flag.Int("timeout", 60, "Client timeout in seconds")
	testTool    = flag.String("tool", "calculator", "Tool to test (calculator, filesearch, cmdexec, searchreplace, screenshot, websearch, webfetch, rag, codeanalysis, patch, linecount, findcallers, funcdef, stats, all)")
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

	// Initialize and run the client
	client := NewClient(*serverURL)
	if err := client.Run(signalCtx, *testTool); err != nil {
		log.Fatalf("Client error: %v", err)
	}

	log.Println("Client operations completed successfully")
}
