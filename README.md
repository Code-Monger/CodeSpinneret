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
│       ├── main.go         # Test client entry point
│       ├── resources.go    # Resource registration
│       └── tools/          # Tool implementations
│           ├── calculator.go
│           ├── cmdexec.go
│           ├── codeanalysis.go
│           ├── filesearch.go
│           ├── patch.go
│           ├── rag.go
│           ├── screenshot.go
│           ├── searchreplace.go
│           ├── spellcheck.go
│           ├── stats.go
│           ├── tools.go
│           ├── webfetch.go
│           └── websearch.go
├── pkg/
│   ├── calculator/         # Calculator tool implementation
│   ├── cmdexec/            # Command execution tool implementation
│   ├── codeanalysis/       # Code analysis tool implementation
│   ├── filesearch/         # File search tool implementation
│   ├── findcallers/        # Find callers tool implementation
│   ├── findfunc/           # Find function tool implementation
│   ├── funcdef/            # Function definition tool implementation
│   ├── linecount/          # Line count tool implementation
│   ├── patch/              # Patch tool implementation
│   ├── rag/                # RAG tool implementation
│   ├── screenshot/         # Screenshot tool implementation
│   ├── searchreplace/      # Search and replace tool implementation
│   ├── serverinfo/         # Server info resource implementation
│   ├── spellcheck/         # Spell check tool implementation with embedded dictionary
│   │   └── data/           # Embedded dictionary data
│   ├── stats/              # Statistics tool implementation
│   ├── test/               # Test utilities
│   ├── webfetch/           # Web fetch tool implementation
│   └── websearch/          # Web search tool implementation
└── bin/                    # Output directory for compiled binaries
    ├── mcp-server.exe      # Compiled server binary
    └── mcp-client.exe      # Compiled client binary
```

## Requirements

- Go 1.23+ (for embed package support)
- Windows (for PowerShell build script)
- Dependencies:
  - github.com/mark3labs/mcp-go: MCP SDK
  - github.com/sajari/fuzzy: Fuzzy matching for spell checking
  - github.com/google/uuid: UUID generation
  - github.com/yosida95/uritemplate/v3: URI template parsing

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

To test a specific tool in demo mode:

```powershell
.\build.ps1 -Target demo -TestTool spellcheck
```

This will start the server, run the client, and test the specified tool with sample data.

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
- `-test-tool`: Specific tool to test in demo mode (used with the demo target)

## Implemented Components

### Tools

- **Calculator**: Performs basic arithmetic operations (add, subtract, multiply, divide)
- **File Search**: Searches for files based on various criteria like name patterns, content, size, and modification time
- **Command Execution**: Executes commands on the system, such as running scripts, compiling code, or starting applications
- **Search Replace**: Finds and replaces text in files, with support for regular expressions and batch operations
- **Screenshot**: Takes screenshots of the screen, windows, or specific regions
- **Web Search**: Performs web searches to find information online, allowing AI models to access up-to-date information
- **Web Fetch**: Fetches the content of a web page given a URL, with options to include/exclude images and set timeout
- **RAG (Retrieval Augmented Generation)**: Provides AI-powered code assistance by retrieving relevant code snippets and generating contextual responses
- **Code Analysis**: Analyzes code to provide insights, metrics, and suggestions for improvement
- **Patch**: Applies patches to files using the standard unified diff format, supporting various options like strip level and dry run
- **Stats**: Tracks and reports usage statistics for all MCP tools, including call counts, execution times, and estimated token savings
- **SpellCheck**: Checks spelling in code comments, string literals, and identifiers. Supports multiple programming languages and can detect misspellings in different naming conventions (camelCase, snake_case, PascalCase). Uses a comprehensive dictionary with over 370,000 words and the sajari/fuzzy package for intelligent suggestion generation ordered by likelihood.
- **FindCallers**: Finds all callers of a specified function across a codebase. Supports multiple programming languages including Go, JavaScript, Python, Java, C#, C/C++, Ruby, and PHP. Returns detailed results with file paths, line numbers, and context for each call.
- **FindFunc**: Finds function definitions across a codebase by name and returns their locations. Supports multiple programming languages and can filter by package name.
- **FuncDef**: Gets or replaces function definitions in source code files across multiple programming languages. Handles complex code patterns including nested functions, comments, and string literals.
- **LineCount**: Counts lines, words, and characters in a file, similar to the Unix 'wc' command. Provides detailed statistics about file content with configurable counting options.

### Resources

- **Server Info**: Provides information about the server (OS, Go version, memory stats, etc.)

## Recent Enhancements

### SpellCheck Tool with Embedded Dictionary and Fuzzy Matching

The SpellCheck tool has been significantly enhanced with the following features:

- **Embedded Dictionary**: Integrated a comprehensive English dictionary with over 370,000 words using Go's embed package, ensuring the dictionary is always available without requiring external files.
- **Fuzzy Matching**: Implemented the sajari/fuzzy package for intelligent suggestion generation, providing more accurate and relevant spelling suggestions ordered by likelihood.
- **Modular Architecture**: Refactored the spellcheck package into multiple logical components for better organization and maintainability:
  - types.go - Contains type definitions and constants
  - languages.go - Contains language-related functionality
  - dictionary.go - Contains word dictionaries and misspelling data
  - utils.go - Contains utility functions
  - checker.go - Contains core spell checking logic
  - handler.go - Contains MCP handler and registration
  - fuzzy.go - Handles the fuzzy matching functionality
- **Multi-language Support**: Added support for multiple programming languages, with language-specific rules for comments, string literals, and identifiers.
- **Naming Convention Detection**: Can detect misspellings in different naming conventions (camelCase, snake_case, PascalCase).
- **Customizable Dictionaries**: Supports custom dictionaries for domain-specific terminology.

### Code Navigation Tools

Several new tools have been added to improve code navigation and analysis:

- **FindCallers**: Finds all callers of a specified function across a codebase, making it easier to understand function usage patterns and perform impact analysis for refactoring.
- **FindFunc**: Locates function definitions across a codebase by name, supporting multiple programming languages and package-based filtering.
- **FuncDef**: Gets or replaces function definitions in source code files, handling complex code patterns including nested functions, comments, and string literals.
- **LineCount**: Provides detailed statistics about file content, similar to the Unix 'wc' command but with more configurable options.

### Patch Tool Enhancements

The Patch tool has been improved to support:

- **Configurable Root Directory**: Added support for a configurable root directory to ensure relative paths work correctly for the target_directory parameter.
- **Consistent Path Handling**: Uses the PATCH_ROOT_DIR environment variable for consistent relative path resolution across tools.
- **Detailed Reporting**: Provides detailed reporting of applied and failed hunks, with dry-run capability for safe testing.

## License

This project is licensed under the MIT License - see the LICENSE file for details.