package workspace

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// WorkspaceInfo represents the workspace information
type WorkspaceInfo struct {
	RootDir    string    `json:"root_dir"`
	UserTask   string    `json:"user_task"`
	InitTime   time.Time `json:"init_time"`
	SessionID  string    `json:"session_id"`
	LastAccess time.Time `json:"last_access"`
}

// SessionStore manages workspace information for multiple sessions
type SessionStore struct {
	sessions map[string]WorkspaceInfo
	mutex    sync.RWMutex
}

// Global session store
var sessionStore = &SessionStore{
	sessions: make(map[string]WorkspaceInfo),
}

// GetWorkspaceInfo returns the workspace info for a session
func GetWorkspaceInfo(sessionID string) (WorkspaceInfo, bool) {
	sessionStore.mutex.RLock()
	defer sessionStore.mutex.RUnlock()

	info, exists := sessionStore.sessions[sessionID]
	if exists {
		// Update last access time
		info.LastAccess = time.Now()
		sessionStore.sessions[sessionID] = info
	}

	return info, exists
}

// SetWorkspaceInfo sets the workspace info for a session
func SetWorkspaceInfo(info WorkspaceInfo) {
	if info.SessionID == "" {
		info.SessionID = fmt.Sprintf("session-%d", time.Now().Unix())
	}

	sessionStore.mutex.Lock()
	defer sessionStore.mutex.Unlock()

	info.InitTime = time.Now()
	info.LastAccess = time.Now()
	sessionStore.sessions[info.SessionID] = info
}

// ListSessions returns a list of all session IDs
func ListSessions() []string {
	sessionStore.mutex.RLock()
	defer sessionStore.mutex.RUnlock()

	sessions := make([]string, 0, len(sessionStore.sessions))
	for sessionID := range sessionStore.sessions {
		sessions = append(sessions, sessionID)
	}

	return sessions
}

// GetRootDir returns the workspace root directory for a session
func GetRootDir(sessionID string) string {
	info, exists := GetWorkspaceInfo(sessionID)
	if !exists {
		return "." // Default to current directory if session doesn't exist
	}

	return info.RootDir
}

// ResolveRelativePath resolves a relative path against the workspace root directory for a session
func ResolveRelativePath(path string, sessionID string) string {
	if filepath.IsAbs(path) {
		return path
	}

	rootDir := GetRootDir(sessionID)
	if rootDir == "" {
		rootDir = "." // Default to current directory if not set
	}

	return filepath.Join(rootDir, path)
}

// HandleWorkspace is the handler function for the workspace tool
func HandleWorkspace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract operation
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	switch operation {
	case "initialize":
		// Extract root directory
		rootDir, ok := arguments["root_dir"].(string)
		if !ok {
			return nil, fmt.Errorf("root_dir must be a string")
		}

		// Extract user task
		userTask, ok := arguments["user_task"].(string)
		if !ok {
			return nil, fmt.Errorf("user_task must be a string")
		}

		// Extract session ID (optional)
		sessionID, _ := arguments["session_id"].(string)
		if sessionID == "" {
			sessionID = fmt.Sprintf("session-%d", time.Now().Unix())
		}

		// Set workspace info
		SetWorkspaceInfo(WorkspaceInfo{
			RootDir:   rootDir,
			UserTask:  userTask,
			SessionID: sessionID,
		})

		// Format the result
		resultText := fmt.Sprintf("Workspace initialized successfully\n\n")
		resultText += fmt.Sprintf("Root directory: %s\n", rootDir)
		resultText += fmt.Sprintf("User task: %s\n", userTask)
		resultText += fmt.Sprintf("Session ID: %s\n", sessionID)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	case "get":
		// Extract session ID
		sessionID, ok := arguments["session_id"].(string)
		if !ok {
			return nil, fmt.Errorf("session_id must be a string")
		}

		// Get workspace info
		info, exists := GetWorkspaceInfo(sessionID)
		if !exists {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}

		// Format the result
		resultText := fmt.Sprintf("Workspace Information\n\n")
		resultText += fmt.Sprintf("Root directory: %s\n", info.RootDir)
		resultText += fmt.Sprintf("User task: %s\n", info.UserTask)
		resultText += fmt.Sprintf("Session ID: %s\n", info.SessionID)
		resultText += fmt.Sprintf("Initialized: %s\n", info.InitTime.Format(time.RFC3339))
		resultText += fmt.Sprintf("Last accessed: %s\n", info.LastAccess.Format(time.RFC3339))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	case "list":
		// List all sessions
		sessions := ListSessions()

		// Format the result
		resultText := fmt.Sprintf("Active Sessions (%d)\n\n", len(sessions))
		for i, sessionID := range sessions {
			info, _ := GetWorkspaceInfo(sessionID)
			resultText += fmt.Sprintf("%d. Session ID: %s\n", i+1, sessionID)
			resultText += fmt.Sprintf("   Root directory: %s\n", info.RootDir)
			resultText += fmt.Sprintf("   User task: %s\n", info.UserTask)
			resultText += fmt.Sprintf("   Initialized: %s\n", info.InitTime.Format(time.RFC3339))
			resultText += fmt.Sprintf("   Last accessed: %s\n\n", info.LastAccess.Format(time.RFC3339))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// HandleWorkspaceResource is the handler function for the workspace resource
func HandleWorkspaceResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// Extract session ID from URI
	uri := request.Params.URI
	sessionID := ""

	// Parse URI to extract session ID
	// Format: workspace://info/session_id
	if strings.HasPrefix(uri, "workspace://info/") {
		sessionID = strings.TrimPrefix(uri, "workspace://info/")
	}

	if sessionID == "" {
		// List all sessions if no session ID is provided
		sessions := ListSessions()

		// Format the result
		infoStr := fmt.Sprintf("Active Sessions (%d):\n\n", len(sessions))
		for i, id := range sessions {
			info, _ := GetWorkspaceInfo(id)
			infoStr += fmt.Sprintf("%d. Session ID: %s\n", i+1, id)
			infoStr += fmt.Sprintf("   Root directory: %s\n", info.RootDir)
			infoStr += fmt.Sprintf("   User task: %s\n", info.UserTask)
			infoStr += fmt.Sprintf("   Initialized: %s\n", info.InitTime.Format(time.RFC3339))
			infoStr += fmt.Sprintf("   Last accessed: %s\n\n", info.LastAccess.Format(time.RFC3339))
		}

		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      request.Params.URI,
				MIMEType: "text/plain",
				Text:     infoStr,
			},
		}, nil
	}

	// Get workspace info for the specified session
	info, exists := GetWorkspaceInfo(sessionID)
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Format the result
	infoStr := fmt.Sprintf("Workspace Information:\n\n")
	infoStr += fmt.Sprintf("root_dir: %s\n", info.RootDir)
	infoStr += fmt.Sprintf("user_task: %s\n", info.UserTask)
	infoStr += fmt.Sprintf("session_id: %s\n", info.SessionID)
	infoStr += fmt.Sprintf("init_time: %s\n", info.InitTime.Format(time.RFC3339))
	infoStr += fmt.Sprintf("last_access: %s\n", info.LastAccess.Format(time.RFC3339))

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "text/plain",
			Text:     infoStr,
		},
	}, nil
}

// RegisterWorkspace registers the workspace tool and resource with the MCP server
func RegisterWorkspace(mcpServer *server.MCPServer) {
	// Register the workspace tool
	workspaceTool := mcp.NewTool("workspace",
		mcp.WithDescription("Initializes and manages the workspace for the model"),
		mcp.WithString("operation",
			mcp.Description("Operation to perform: 'initialize' to set up the workspace, 'get' to retrieve workspace information, 'list' to list all sessions"),
			mcp.Required(),
		),
		mcp.WithString("root_dir",
			mcp.Description("Root directory of the source code (for 'initialize' operation)"),
		),
		mcp.WithString("user_task",
			mcp.Description("Task the user has set for the model (for 'initialize' operation)"),
		),
		mcp.WithString("session_id",
			mcp.Description("Session ID (required for 'get' operation, optional for 'initialize' operation)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("workspace", HandleWorkspace)

	// Register the tool
	mcpServer.AddTool(workspaceTool, wrappedHandler)

	// Register the workspace resource
	mcpServer.AddResource(
		mcp.NewResource(
			"workspace://info",
			"Workspace Information",
			mcp.WithMIMEType("text/plain"),
		),
		HandleWorkspaceResource,
	)

	// Register the workspace resource template for session-specific URIs
	mcpServer.AddResourceTemplate(
		mcp.NewResourceTemplate(
			"workspace://info/{session_id}",
			"Workspace Session Information",
			mcp.WithTemplateMIMEType("text/plain"),
			mcp.WithTemplateDescription("Information about a specific workspace session"),
		),
		HandleWorkspaceResource,
	)

	// Log the registration
	log.Printf("[Workspace] Registered workspace tool and resource")
}
