package findfunc

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

// Language represents a programming language with its file extensions and function definition patterns
type Language struct {
	Name                 string
	FileExtensions       []string
	FunctionStartPattern string
}

// FunctionLocation represents the location of a function in a file
type FunctionLocation struct {
	FilePath    string `json:"file_path"`
	LineNumber  int    `json:"line_number"`
	PackageName string `json:"package_name,omitempty"`
	Language    string `json:"language"`
}

// GetSupportedLanguages returns a list of supported programming languages
func GetSupportedLanguages() []Language {
	return []Language{
		{
			Name:                 "Go",
			FileExtensions:       []string{".go"},
			FunctionStartPattern: `func\s+%s\s*\(`,
		},
		{
			Name:                 "JavaScript",
			FileExtensions:       []string{".js", ".jsx", ".ts", ".tsx"},
			FunctionStartPattern: `(function\s+%s|const\s+%s\s*=\s*function|let\s+%s\s*=\s*function|var\s+%s\s*=\s*function|%s\s*=\s*function|%s\s*:\s*function|%s\s*\([^{]*\)\s*{|%s\s*=\s*\([^{]*\)\s*=>|%s\s*=>)`,
		},
		{
			Name:                 "Python",
			FileExtensions:       []string{".py"},
			FunctionStartPattern: `def\s+%s\s*\(`,
		},
		{
			Name:                 "Java",
			FileExtensions:       []string{".java"},
			FunctionStartPattern: `(public|private|protected)?\s*(static)?\s*\S+\s+%s\s*\(`,
		},
		{
			Name:                 "C#",
			FileExtensions:       []string{".cs"},
			FunctionStartPattern: `(public|private|protected|internal)?\s*(static|virtual|override|abstract)?\s*\S+\s+%s\s*\(`,
		},
		{
			Name:                 "C/C++",
			FileExtensions:       []string{".c", ".cpp", ".cc", ".h", ".hpp"},
			FunctionStartPattern: `\S+\s+%s\s*\(`,
		},
		{
			Name:                 "Ruby",
			FileExtensions:       []string{".rb"},
			FunctionStartPattern: `(def\s+%s|def\s+self\.%s)`,
		},
		{
			Name:                 "PHP",
			FileExtensions:       []string{".php"},
			FunctionStartPattern: `(function\s+%s|public\s+function\s+%s|private\s+function\s+%s|protected\s+function\s+%s|static\s+function\s+%s)`,
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
			if strings.EqualFold(langExt, ext) {
				return lang, true
			}
		}
	}
	return Language{}, false
}

// HandleFindFunc is the handler function for the findfunc tool
func HandleFindFunc(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract function name
	functionName, ok := arguments["function_name"].(string)
	if !ok {
		return nil, fmt.Errorf("function_name must be a string")
	}

	// Extract package name (optional)
	packageName, _ := arguments["package_name"].(string)

	// Extract search directory
	searchDir, _ := arguments["search_directory"].(string)
	if searchDir == "" {
		searchDir = "." // Default to current directory
	}

	// Get root directory from environment variable
	rootDir := os.Getenv(EnvRootDir)
	if rootDir == "" {
		rootDir = searchDir // Default to search directory if env var not set
	}

	// Resolve the search directory path
	var fullSearchDir string
	if filepath.IsAbs(searchDir) {
		fullSearchDir = searchDir
	} else {
		fullSearchDir = filepath.Join(rootDir, searchDir)
	}

	// Extract language (optional)
	language, _ := arguments["language"].(string)

	// Extract recursive flag
	recursive := true // Default to recursive search
	if recursiveBool, ok := arguments["recursive"].(bool); ok {
		recursive = recursiveBool
	}

	// Extract use_relative_paths flag
	useRelativePaths := true // Default to using relative paths
	if useRelativePathsBool, ok := arguments["use_relative_paths"].(bool); ok {
		useRelativePaths = useRelativePathsBool
	}

	// Find functions
	locations, err := findFunctions(fullSearchDir, functionName, packageName, language, recursive)
	if err != nil {
		return nil, fmt.Errorf("error finding functions: %v", err)
	}

	// Convert paths to relative if requested
	if useRelativePaths {
		for i := range locations {
			relPath, err := filepath.Rel(rootDir, locations[i].FilePath)
			if err == nil {
				locations[i].FilePath = relPath
			}
		}
	}

	// Create the result
	result := &mcp.CallToolResult{}

	// Add the locations to the result
	if len(locations) > 0 {
		// Create a text summary
		var summary strings.Builder
		summary.WriteString(fmt.Sprintf("Found %d function(s) matching '%s':\n\n", len(locations), functionName))

		for i, loc := range locations {
			summary.WriteString(fmt.Sprintf("%d. File: %s\n", i+1, loc.FilePath))
			summary.WriteString(fmt.Sprintf("   Line: %d\n", loc.LineNumber))
			if loc.PackageName != "" {
				summary.WriteString(fmt.Sprintf("   Package: %s\n", loc.PackageName))
			}
			summary.WriteString(fmt.Sprintf("   Language: %s\n", loc.Language))
			summary.WriteString("\n")
		}

		// Just add the text summary
		result.Content = append(result.Content, mcp.TextContent{
			Text: summary.String(),
			Type: "text",
		})
	} else {
		result.Content = append(result.Content, mcp.TextContent{
			Text: fmt.Sprintf("No functions found matching '%s'", functionName),
			Type: "text",
		})
	}

	return result, nil
}

// findFunctions finds all functions with the given name in the specified directory
func findFunctions(searchDir, functionName, packageName, language string, recursive bool) ([]FunctionLocation, error) {
	var locations []FunctionLocation

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

		// Search for function definitions in the file
		fileLocations, err := searchFileForFunctions(path, functionName, packageName, lang)
		if err != nil {
			log.Printf("[FindFunc] Error searching file %s: %v", path, err)
			return nil
		}

		locations = append(locations, fileLocations...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return locations, nil
}

// searchFileForFunctions searches for function definitions in a file
func searchFileForFunctions(filePath, functionName, packageName string, language Language) ([]FunctionLocation, error) {
	var locations []FunctionLocation

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Create the function pattern
	var functionPattern string
	if strings.Contains(language.FunctionStartPattern, "%s") {
		functionPattern = fmt.Sprintf(language.FunctionStartPattern, functionName)
	} else {
		functionPattern = language.FunctionStartPattern
	}

	// Compile the pattern
	re, err := regexp.Compile(functionPattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling pattern: %v", err)
	}

	// Variables to track package context
	var currentPackage string
	packageRe := regexp.MustCompile(`package\s+(\w+)`)

	// Scan the file line by line
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check for package declaration (for Go)
		if language.Name == "Go" {
			if matches := packageRe.FindStringSubmatch(line); len(matches) > 1 {
				currentPackage = matches[1]
			}
		}

		// Check if the line contains a function definition
		if re.MatchString(line) {
			// If package name is specified, check if it matches
			if packageName != "" && currentPackage != "" && packageName != currentPackage {
				continue
			}

			// Add the location
			locations = append(locations, FunctionLocation{
				FilePath:    filePath,
				LineNumber:  lineNumber,
				PackageName: currentPackage,
				Language:    language.Name,
			})
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return locations, nil
}

// RegisterFindFunc registers the findfunc tool with the MCP server
func RegisterFindFunc(mcpServer *server.MCPServer) {
	// Create the tool definition
	findFuncTool := mcp.NewTool("findfunc",
		mcp.WithDescription("Finds function definitions across a codebase by name and returns their locations. Supports multiple programming languages including Go, JavaScript, Python, Java, C#, C/C++, Ruby, and PHP. Returns an array of locations with file paths and line numbers, making it ideal for use with the funcdef tool to retrieve specific function definitions. Handles package-based languages like Go and Java with optional package name filtering."),
		mcp.WithString("function_name",
			mcp.Description("The name of the function to find (case-sensitive, must match exactly as defined in code)"),
			mcp.Required(),
		),
		mcp.WithString("package_name",
			mcp.Description("The package name to filter by (for languages like Go and Java that use package-based identifiers)"),
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
	wrappedHandler := stats.WrapHandler("findfunc", HandleFindFunc)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(findFuncTool, wrappedHandler)

	log.Printf("[FindFunc] Registered findfunc tool")
}
