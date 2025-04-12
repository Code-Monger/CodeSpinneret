package shell

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/Code-Monger/CodeSpinneret/pkg/workspace"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ShellSession represents a persistent shell session
type ShellSession struct {
	ShellType   string    // Type of shell (bash, powershell, cmd)
	Cmd         *exec.Cmd // The command process
	Stdin       io.WriteCloser
	StdoutPipe  io.ReadCloser
	StderrPipe  io.ReadCloser
	OutputChan  chan string
	ErrChan     chan string
	LastOutput  string
	LastError   string
	LastCommand string
	StartTime   time.Time
	LastAccess  time.Time
	Mutex       sync.Mutex
	Closed      bool
}

// ShellStore manages shell sessions for multiple workspace sessions
type ShellStore struct {
	sessions map[string]*ShellSession
	mutex    sync.RWMutex
}

// Global shell store
var shellStore = &ShellStore{
	sessions: make(map[string]*ShellSession),
}

// GetShellSession returns the shell session for a workspace session
func GetShellSession(sessionID string) (*ShellSession, bool) {
	shellStore.mutex.RLock()
	defer shellStore.mutex.RUnlock()

	session, exists := shellStore.sessions[sessionID]
	if exists {
		// Update last access time
		session.LastAccess = time.Now()
	}

	return session, exists
}

// SetShellSession sets the shell session for a workspace session
func SetShellSession(sessionID string, session *ShellSession) {
	shellStore.mutex.Lock()
	defer shellStore.mutex.Unlock()

	shellStore.sessions[sessionID] = session
}

// RemoveShellSession removes a shell session
func RemoveShellSession(sessionID string) {
	shellStore.mutex.Lock()
	defer shellStore.mutex.Unlock()

	if session, exists := shellStore.sessions[sessionID]; exists {
		// Close the session if it's still open
		if !session.Closed {
			session.Close()
		}
		delete(shellStore.sessions, sessionID)
	}
}

// Close closes the shell session
func (s *ShellSession) Close() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if s.Closed {
		return
	}

	// Send exit command to gracefully terminate the shell
	if s.Stdin != nil {
		s.Stdin.Write([]byte("exit\n"))
		s.Stdin.Close()
	}

	// Close pipes
	if s.StdoutPipe != nil {
		s.StdoutPipe.Close()
	}
	if s.StderrPipe != nil {
		s.StderrPipe.Close()
	}

	// Wait for process to terminate with timeout
	done := make(chan error, 1)
	go func() {
		if s.Cmd != nil && s.Cmd.Process != nil {
			done <- s.Cmd.Wait()
		} else {
			done <- nil
		}
	}()

	// Wait with timeout
	select {
	case <-done:
		// Process exited normally
	case <-time.After(2 * time.Second):
		// Force kill if it doesn't exit gracefully
		if s.Cmd != nil && s.Cmd.Process != nil {
			s.Cmd.Process.Kill()
		}
	}

	s.Closed = true
}

// InitializeShell initializes a new shell session for a workspace session
func InitializeShell(sessionID string, shellType string) (*ShellSession, error) {
	// Check if a session already exists
	if existingSession, exists := GetShellSession(sessionID); exists {
		// Close the existing session
		existingSession.Close()
	}

	// Determine the shell command based on the shell type and OS
	var cmd *exec.Cmd
	if shellType == "" {
		// Default shell based on OS
		if runtime.GOOS == "windows" {
			shellType = "cmd"
		} else {
			shellType = "bash"
		}
	}

	switch shellType {
	case "bash":
		cmd = exec.Command("bash")
	case "powershell":
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-NoExit", "-Command", "-")
		} else {
			cmd = exec.Command("pwsh", "-NoExit", "-Command", "-")
		}
	case "cmd":
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd")
		} else {
			return nil, fmt.Errorf("cmd shell is only available on Windows")
		}
	default:
		return nil, fmt.Errorf("unsupported shell type: %s", shellType)
	}

	// Get workspace info to set working directory
	workspaceInfo, exists := workspace.GetWorkspaceInfo(sessionID)
	if exists {
		cmd.Dir = workspaceInfo.RootDir
	}

	// Create pipes for stdin, stdout, and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Create the shell session
	session := &ShellSession{
		ShellType:  shellType,
		Cmd:        cmd,
		Stdin:      stdin,
		StdoutPipe: stdout,
		StderrPipe: stderr,
		OutputChan: make(chan string, 100),
		ErrChan:    make(chan string, 100),
		StartTime:  time.Now(),
		LastAccess: time.Now(),
		Closed:     false,
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to start shell: %v", err)
	}

	// Start goroutines to read stdout and stderr
	go readOutput(stdout, session.OutputChan, &session.LastOutput, &session.Mutex)
	go readOutput(stderr, session.ErrChan, &session.LastError, &session.Mutex)

	// Store the session
	SetShellSession(sessionID, session)

	return session, nil
}

// readOutput reads from a pipe and sends the output to a channel
func readOutput(pipe io.Reader, outputChan chan string, lastOutput *string, mutex *sync.Mutex) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		outputChan <- line

		// Update last output
		mutex.Lock()
		*lastOutput += line + "\n"
		mutex.Unlock()
	}
}

// ExecuteCommand executes a command in the shell session
func ExecuteCommand(session *ShellSession, command string, timeoutSec float64) (string, string, error) {
	session.Mutex.Lock()
	if session.Closed {
		session.Mutex.Unlock()
		return "", "", fmt.Errorf("shell session is closed")
	}

	// Clear previous output
	session.LastOutput = ""
	session.LastError = ""
	session.LastCommand = command
	session.Mutex.Unlock()

	// Write the command to stdin
	if _, err := session.Stdin.Write([]byte(command + "\n")); err != nil {
		return "", "", fmt.Errorf("failed to write command to shell: %v", err)
	}

	// Wait for output with timeout
	timeout := time.After(time.Duration(timeoutSec) * time.Second)
	outputDone := false
	errorDone := false

	// Collect output until timeout or both channels are empty
	for !outputDone || !errorDone {
		select {
		case <-timeout:
			return session.LastOutput, session.LastError, nil
		case output, ok := <-session.OutputChan:
			if !ok {
				outputDone = true
			} else {
				session.Mutex.Lock()
				session.LastOutput += output + "\n"
				session.Mutex.Unlock()
			}
		case errOutput, ok := <-session.ErrChan:
			if !ok {
				errorDone = true
			} else {
				session.Mutex.Lock()
				session.LastError += errOutput + "\n"
				session.Mutex.Unlock()
			}
		case <-time.After(100 * time.Millisecond):
			// Check if there's been no output for a short time
			// This helps detect when the command has finished
			select {
			case output, ok := <-session.OutputChan:
				if !ok {
					outputDone = true
				} else {
					session.Mutex.Lock()
					session.LastOutput += output + "\n"
					session.Mutex.Unlock()
				}
			case errOutput, ok := <-session.ErrChan:
				if !ok {
					errorDone = true
				} else {
					session.Mutex.Lock()
					session.LastError += errOutput + "\n"
					session.Mutex.Unlock()
				}
			default:
				// No output for 100ms, assume command is done
				return session.LastOutput, session.LastError, nil
			}
		}
	}

	return session.LastOutput, session.LastError, nil
}

// HandleShell is the handler function for the shell tool
func HandleShell(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract operation
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	// Extract session ID
	sessionID, ok := arguments["session_id"].(string)
	if !ok {
		return nil, fmt.Errorf("session_id must be a string")
	}

	switch operation {
	case "initialize":
		// Extract shell type (optional)
		shellType, _ := arguments["shell_type"].(string)

		// Initialize the shell
		session, err := InitializeShell(sessionID, shellType)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize shell: %v", err)
		}

		// Format the result
		resultText := fmt.Sprintf("Shell initialized successfully\n\n")
		resultText += fmt.Sprintf("Session ID: %s\n", sessionID)
		resultText += fmt.Sprintf("Shell Type: %s\n", session.ShellType)
		resultText += fmt.Sprintf("Started: %s\n", session.StartTime.Format(time.RFC3339))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	case "execute":
		// Extract command
		command, ok := arguments["command"].(string)
		if !ok {
			return nil, fmt.Errorf("command must be a string")
		}

		// Extract timeout (optional)
		timeoutSec, ok := arguments["timeout"].(float64)
		if !ok {
			// Default timeout: 30 seconds
			timeoutSec = 30
		}

		// Get the shell session
		session, exists := GetShellSession(sessionID)
		if !exists {
			// Try to initialize a new session if it doesn't exist
			var err error
			shellType, _ := arguments["shell_type"].(string)
			session, err = InitializeShell(sessionID, shellType)
			if err != nil {
				return nil, fmt.Errorf("shell session not found and failed to initialize: %v", err)
			}
		}

		// Execute the command
		stdout, stderr, err := ExecuteCommand(session, command, timeoutSec)
		if err != nil {
			return nil, fmt.Errorf("failed to execute command: %v", err)
		}

		// Format the result
		resultText := fmt.Sprintf("Command executed in shell session\n\n")
		resultText += fmt.Sprintf("Session ID: %s\n", sessionID)
		resultText += fmt.Sprintf("Shell Type: %s\n", session.ShellType)
		resultText += fmt.Sprintf("Command: %s\n\n", command)

		if stdout != "" {
			resultText += fmt.Sprintf("Standard Output:\n%s\n", stdout)
		}
		if stderr != "" {
			resultText += fmt.Sprintf("Standard Error:\n%s\n", stderr)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	case "close":
		// Close the shell session
		session, exists := GetShellSession(sessionID)
		if !exists {
			return nil, fmt.Errorf("shell session not found: %s", sessionID)
		}

		// Close the session
		session.Close()
		RemoveShellSession(sessionID)

		// Format the result
		resultText := fmt.Sprintf("Shell session closed\n\n")
		resultText += fmt.Sprintf("Session ID: %s\n", sessionID)
		resultText += fmt.Sprintf("Shell Type: %s\n", session.ShellType)
		resultText += fmt.Sprintf("Duration: %v\n", time.Since(session.StartTime))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: resultText,
				},
			},
		}, nil

	case "status":
		// Get the shell session status
		session, exists := GetShellSession(sessionID)
		if !exists {
			return nil, fmt.Errorf("shell session not found: %s", sessionID)
		}

		// Format the result
		resultText := fmt.Sprintf("Shell Session Status\n\n")
		resultText += fmt.Sprintf("Session ID: %s\n", sessionID)
		resultText += fmt.Sprintf("Shell Type: %s\n", session.ShellType)
		resultText += fmt.Sprintf("Started: %s\n", session.StartTime.Format(time.RFC3339))
		resultText += fmt.Sprintf("Last Access: %s\n", session.LastAccess.Format(time.RFC3339))
		resultText += fmt.Sprintf("Duration: %v\n", time.Since(session.StartTime))
		resultText += fmt.Sprintf("Last Command: %s\n", session.LastCommand)
		resultText += fmt.Sprintf("Status: %s\n", func() string {
			if session.Closed {
				return "Closed"
			}
			return "Active"
		}())

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

// RegisterShell registers the shell tool with the MCP server
func RegisterShell(mcpServer *server.MCPServer) {
	// Create the tool definition
	shellTool := mcp.NewTool("shell",
		mcp.WithDescription("Maintains a persistent shell session for executing commands with context"),
		mcp.WithString("operation",
			mcp.Description("Operation to perform: 'initialize', 'execute', 'close', or 'status'"),
			mcp.Required(),
		),
		mcp.WithString("session_id",
			mcp.Description("Session ID to associate with the shell"),
			mcp.Required(),
		),
		mcp.WithString("shell_type",
			mcp.Description("Type of shell to use: 'bash', 'powershell', or 'cmd' (default depends on OS)"),
		),
		mcp.WithString("command",
			mcp.Description("The command to execute (for 'execute' operation)"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Timeout in seconds for command execution (default: 30)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("shell", HandleShell)

	// Register the tool
	mcpServer.AddTool(shellTool, wrappedHandler)

	// Log the registration
	log.Printf("[Shell] Registered shell tool")
}
