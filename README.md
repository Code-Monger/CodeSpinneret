# CodeSpinneret MCP Server

A Model Context Protocol (MCP) server implementation using the [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) SDK.

## Overview

This project implements a server and test client that follow the Model Context Protocol, allowing AI models to interact with external tools and data sources through a standardized interface.

## Features

- **Server-Sent Events (SSE)** for real-time communication
- **Configurable Timeouts** for both server and client
- **Example Tool**: Calculator for basic arithmetic operations
- **Example Resource**: Server information resource
- **Graceful Shutdown**: Proper handling of termination signals

## Project Structure

```
CodeSpinneret/
├── build.ps1               # PowerShell build script
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies checksum
├── README.md               # Project documentation
├── cmd/
│   ├── mcp-server/
│   │   └── main.go         # Server entry point
│   └── mcp-client/
│       └── main.go         # Test client entry point
├── pkg/
│   ├── calculator/
│   │   └── calculator.go   # Calculator tool implementation
│   └── serverinfo/
│       └── serverinfo.go   # Server info resource implementation
└── bin/                    # Output directory for compiled binaries
    ├── mcp-server.exe      # Compiled server binary
    └── mcp-client.exe      # Compiled client binary
```

## Requirements

- Go 1.21+
- Windows (for PowerShell build script)

## Building and Running

The project includes a PowerShell build script (`build.ps1`) that provides various targets for building, testing, and running the server and client.

### Build

To build both the server and client:

```powershell
.\build.ps1 -Target build
```

### Run Server

To run the server:

```powershell
.\build.ps1 -Target run-server -Port 8080
```

### Run Client

To run the client (connecting to a running server):

```powershell
.\build.ps1 -Target run-client -Port 8080
```

### Demo Mode

To run both server and client in demo mode:

```powershell
.\build.ps1 -Target demo -Port 8080
```

### Other Targets

- `build-server`: Build only the server
- `build-client`: Build only the client
- `test`: Run tests
- `clean`: Clean build artifacts
- `update-deps`: Update dependencies

## Configuration

Both the server and client support various command-line flags for configuration:

### Server Flags

- `-port`: Port to listen on (default: 8080)
- `-baseurl`: Base URL for the server (e.g., http://localhost:8080)
- `-name`: Server name (default: "CodeSpinneret MCP Server")
- `-version`: Server version (default: "1.0.0")
- `-timeout`: Server timeout in seconds (default: 300)
- `-instructions`: Server instructions

### Client Flags

- `-server`: MCP server URL (default: "http://localhost:8080")
- `-timeout`: Client timeout in seconds (default: 60)
- `-tool`: Tool to test (default: "calculator")

## Implemented Components

### Tools

- **Calculator**: Performs basic arithmetic operations (add, subtract, multiply, divide)
- **File Search**: Searches for files based on various criteria like name patterns, content, size, and modification time
- **Command Execution**: Executes commands on the system, such as running scripts, compiling code, or starting applications
- **Search Replace**: Finds and replaces text in files, with support for regular expressions and batch operations
- **Screenshot**: Takes screenshots of the screen, windows, or specific regions
- **Web Search**: Performs web searches to find information online, allowing AI models to access up-to-date information
- **RAG (Retrieval Augmented Generation)**: Provides AI-powered code assistance by retrieving relevant code snippets and generating contextual responses
- **Code Analysis**: Analyzes code to provide insights, metrics, and suggestions for improvement

### Resources

- **Server Info**: Provides information about the server (OS, Go version, memory stats, etc.)

## License

This project is licensed under the MIT License - see the LICENSE file for details.