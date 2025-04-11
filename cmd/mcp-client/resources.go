package main

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// ReadServerInfo reads the server info resource
func ReadServerInfo(ctx context.Context, c client.MCPClient) error {
	readReq := mcp.ReadResourceRequest{}
	readReq.Params.URI = "server://info"

	result, err := c.ReadResource(ctx, readReq)
	if err != nil {
		log.Printf("Failed to read server info: %v", err)
		return err
	}

	if len(result.Contents) > 0 {
		if textContent, ok := result.Contents[0].(mcp.TextResourceContents); ok {
			log.Printf("Server Info:\n%s", textContent.Text)
		}
	}

	return nil
}
