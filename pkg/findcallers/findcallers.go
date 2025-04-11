package findcallers

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Environment variable name for root directory (same as patch and linecount tools)
const EnvRootDir = "PATCH_ROOT_DIR"

// Language represents a programming language with its file extensions and function call patterns
type Language struct {
	Name           string
	FileExtensions []string
	CallPatterns   []string
}

// GetSupportedLanguages returns a list of supported programming languages
func GetSupportedLanguages() []Language {
	return []Language{
		{
			Name:           "Go",
			FileExtensions: []string{".go"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*\(`},
		},
		{
			Name:           "JavaScript",
			FileExtensions: []string{".js", ".jsx", ".ts", ".tsx"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*\(`, `(^|\s+|\.)%s\s*\(`},
		},
		{
			Name:           "Python",
			FileExtensions: []string{".py"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*\(`},
		},
		{
			Name:           "Java",
			FileExtensions: []string{".java"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*\(`, `(^|\s+|\.)%s\s*\(`},
		},
		{
			Name:           "C#",
			FileExtensions: []string{".cs"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*\(`, `(^|\s+|\.)%s\s*\(`},
		},
		{
			Name:           "C/C++",
			FileExtensions: []string{".c", ".cpp", ".cc", ".h", ".hpp"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*\(`, `(^|\s+|::|->)%s\s*\(`},
		},
		{
			Name:           "Ruby",
			FileExtensions: []string{".rb"},
			CallPatterns:   []string{`(^|\s+|[^\w\.])%s\s*(\(|$)`, `(^|\s+|\.)%s\s*(\(|$)`},
		},
		{
			Name:           "PHP",
			FileExtensions: []string{".php"},
			CallPatterns:   []string{`(^|\s+|[^\w\$\.])%s\s*\(`, `(^|\s+|\$|\->|\:\:)%s\s*\(`},
		},
	}
}

// GetLanguageByName returns a language by its name
func GetLanguageByName(name string) (Language, bool) {
	languages := GetSupportedLanguages()
	for _, lang := range languages {
		if strings.EqualFold(lang.Name, name) {
			return lang, true
		}
	}
	return Language{}, false
}

// GetLanguageByExtension returns a language by file extension
func GetLanguageByExtension(ext string) (Language, bool) {
	languages := GetSupportedLanguages()
	for _, lang := range languages {
		for _, langExt := range lang.FileExtensions {
			if langExt == ext {
				return lang, true
			}
		}
	}
	return Language{}, false
}

// CallerResult represents a found caller of a function
type CallerResult struct {
	FilePath    string
	LineNumber  int
	LineContent string
	Language    string
}

// HandleFindCallers is the handler function for the findcallers tool
func HandleFindCallers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract function name
	functionName, ok := arguments["function_name"].(string)
	if !ok {
		return nil, fmt.Errorf("function_name must be a string")
	}

	// Extract search directory
	searchDir, ok := arguments["search_directory"].(string)
	if !ok {
		searchDir = "."
	}

	// Extract language
	language := ""
	if langVal, ok := arguments["language"].(string); ok {
		language = langVal
	}

	// Extract use_relative_paths flag
	useRelativePaths := true
	if relPathsVal, ok := arguments["use_relative_paths"].(bool); ok {
		useRelativePaths = relPathsVal
	}

	// Extract recursive flag
	recursive := true
	if recursiveVal, ok := arguments["recursive"].(bool); ok {
		recursive = recursiveVal
	}

	// Get root directory from environment variable
	rootDir := os.Getenv(EnvRootDir)
	if rootDir == "" {
		rootDir = "." // Default to current directory if env var not set
	}

	// Resolve the search directory path
	var searchDirPath string
	if filepath.IsAbs(searchDir) {
		searchDirPath = searchDir
	} else {
		searchDirPath = filepath.Join(rootDir, searchDir)
	}

	// Log the paths for debugging
	log.Printf("[FindCallers] Function name: %s", functionName)
	log.Printf("[FindCallers] Search directory: %s", searchDirPath)
	log.Printf("[FindCallers] Language: %s", language)
	log.Printf("[FindCallers] Use relative paths: %v", useRelativePaths)
	log.Printf("[FindCallers] Recursive: %v", recursive)

	// Find callers
	results, err := FindCallers(functionName, searchDirPath, language, recursive)
	if err != nil {
		return nil, fmt.Errorf("error finding callers: %v", err)
	}

	// Format the results
	resultText := fmt.Sprintf("Callers of function '%s':\n\n", functionName)

	if len(results) == 0 {
		resultText += "No callers found.\n"
	} else {
		for i, result := range results {
			// Format the file path based on the useRelativePaths flag
			filePath := result.FilePath
			if useRelativePaths {
				relPath, err := filepath.Rel(rootDir, result.FilePath)
				if err == nil {
					filePath = relPath
				}
			}

			resultText += fmt.Sprintf("%d. %s:%d [%s]\n", i+1, filePath, result.LineNumber, result.Language)
			resultText += fmt.Sprintf("   %s\n\n", strings.TrimSpace(result.LineContent))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: resultText,
			},
		},
	}, nil
}

// FindCallers finds all callers of a function in a directory
func FindCallers(functionName, searchDir, language string, recursive bool) ([]CallerResult, error) {
	var results []CallerResult

	// Get supported languages
	supportedLanguages := GetSupportedLanguages()

	// Filter languages if specified
	var languages []Language
	if language != "" {
		lang, found := GetLanguageByName(language)
		if found {
			languages = []Language{lang}
		} else {
			return nil, fmt.Errorf("unsupported language: %s", language)
		}
	} else {
		languages = supportedLanguages
	}

	// Create a map of file extensions to languages
	extToLang := make(map[string]Language)
	for _, lang := range languages {
		for _, ext := range lang.FileExtensions {
			extToLang[ext] = lang
		}
	}

	// Walk the directory
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories if not recursive
		if info.IsDir() {
			if path != searchDir && !recursive {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if the file extension is supported
		ext := filepath.Ext(path)
		lang, ok := extToLang[ext]
		if !ok {
			return nil
		}

		// Search for function calls in the file
		fileResults, err := searchFileForCalls(path, functionName, lang)
		if err != nil {
			log.Printf("[FindCallers] Error searching file %s: %v", path, err)
			return nil
		}

		results = append(results, fileResults...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return results, nil
}

// searchFileForCalls searches a file for function calls
func searchFileForCalls(filePath, functionName string, language Language) ([]CallerResult, error) {
	var results []CallerResult

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Compile regex patterns for the language
	var patterns []*regexp.Regexp
	for _, pattern := range language.CallPatterns {
		regexPattern := fmt.Sprintf(pattern, regexp.QuoteMeta(functionName))
		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex: %v", err)
		}
		patterns = append(patterns, regex)
	}

	// Search for function calls
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check if the line contains a function call
		for _, regex := range patterns {
			if regex.MatchString(line) {
				results = append(results, CallerResult{
					FilePath:    filePath,
					LineNumber:  lineNumber,
					LineContent: line,
					Language:    language.Name,
				})
				break
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return results, nil
}

// RegisterFindCallers registers the findcallers tool with the MCP server
func RegisterFindCallers(mcpServer *server.MCPServer) {
	// Create the tool definition
	findCallersTool := mcp.NewTool("findcallers",
		mcp.WithDescription("Finds all callers of a specified function across a codebase. Supports multiple programming languages including Go, JavaScript, Python, Java, C#, C/C++, Ruby, and PHP. Analyzes code to identify function calls while handling complex patterns like method calls, nested functions, and calls within comments or string literals. Returns detailed results with file paths, line numbers, and context for each call, making it ideal for code refactoring, impact analysis, and understanding function usage patterns."),
		mcp.WithString("function_name",
			mcp.Description("The name of the function to find callers for (case-sensitive, must match exactly as defined in code)"),
			mcp.Required(),
		),
		mcp.WithString("search_directory",
			mcp.Description("The directory to search in (absolute or relative path, default: current directory)"),
		),
		mcp.WithString("language",
			mcp.Description("The programming language to search for (default: all supported languages - Go, JavaScript, Python, Java, C#, C/C++, Ruby, PHP)"),
		),
		mcp.WithBoolean("use_relative_paths",
			mcp.Description("Whether to use relative paths in the results for better portability (default: true)"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to search recursively in subdirectories for comprehensive analysis (default: true)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("findcallers", HandleFindCallers)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(findCallersTool, wrappedHandler)

	log.Printf("[FindCallers] Registered findcallers tool")
}
