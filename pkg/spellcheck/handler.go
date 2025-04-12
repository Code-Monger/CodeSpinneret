package spellcheck

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/Code-Monger/CodeSpinneret/pkg/workspace"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// spellCheckDirectory performs spell checking on all files in a directory
func spellCheckDirectory(dirPath, language string, recursive, checkComments, checkStrings, checkIdentifiers bool, dictionaryType string, customDictionary []string) ([]SpellCheckResult, error) {
	var results []SpellCheckResult

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
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories if not recursive
		if info.IsDir() {
			if path != dirPath && !recursive {
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

		// Spell check the file
		fileResults, err := spellCheckFile(path, lang.Name, checkComments, checkStrings, checkIdentifiers, dictionaryType, customDictionary)
		if err != nil {
			log.Printf("[SpellCheck] Error checking file %s: %v", path, err)
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

// HandleSpellCheck is the handler function for the spellcheck tool
func HandleSpellCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract file or directory path
	path, ok := arguments["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path must be a string")
	}

	// Extract check types
	checkComments := true
	if checkCommentsVal, ok := arguments["check_comments"].(bool); ok {
		checkComments = checkCommentsVal
	}

	checkStrings := true
	if checkStringsVal, ok := arguments["check_strings"].(bool); ok {
		checkStrings = checkStringsVal
	}

	checkIdentifiers := true
	if checkIdentifiersVal, ok := arguments["check_identifiers"].(bool); ok {
		checkIdentifiers = checkIdentifiersVal
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

	// Extract dictionary type
	dictionaryType := "standard" // Default to standard dictionary
	if dictTypeStr, ok := arguments["dictionary_type"].(string); ok {
		dictionaryType = dictTypeStr
	}

	// Extract custom dictionary words
	var customDictionary []string
	if customDictVal, ok := arguments["custom_dictionary"].([]interface{}); ok {
		for _, word := range customDictVal {
			if wordStr, ok := word.(string); ok {
				customDictionary = append(customDictionary, wordStr)
			}
		}
	}
	// Extract session ID
	sessionID, _ := arguments["session_id"].(string)

	// Get root directory from workspace
	rootDir := workspace.GetRootDir(sessionID)

	// Log the root directory for debugging
	log.Printf("[SpellCheck] Using workspace root directory: %s", rootDir)

	// Resolve the path
	var fullPath string
	if filepath.IsAbs(path) {
		fullPath = path
	} else {
		fullPath = filepath.Join(rootDir, path)
	}

	// Check if the path is a file or directory
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("error accessing path: %v", err)
	}

	var results []SpellCheckResult
	if fileInfo.IsDir() {
		// Spell check a directory
		results, err = spellCheckDirectory(fullPath, language, recursive, checkComments, checkStrings, checkIdentifiers, dictionaryType, customDictionary)
	} else {
		// Spell check a single file
		results, err = spellCheckFile(fullPath, language, checkComments, checkStrings, checkIdentifiers, dictionaryType, customDictionary)
	}

	if err != nil {
		return nil, fmt.Errorf("error performing spell check: %v", err)
	}

	// Convert paths to relative if requested
	if useRelativePaths {
		for i := range results {
			relPath, err := filepath.Rel(rootDir, results[i].FilePath)
			if err == nil {
				results[i].FilePath = relPath
			}
		}
	}

	// Create the result
	result := &mcp.CallToolResult{}

	// Add the results to the response
	if len(results) > 0 {
		// Create a text summary
		var summary strings.Builder
		summary.WriteString(fmt.Sprintf("Found %d spelling issues:\n\n", len(results)))

		for i, issue := range results {
			summary.WriteString(fmt.Sprintf("%d. File: %s\n", i+1, issue.FilePath))
			summary.WriteString(fmt.Sprintf("   Line: %d, Columns: %d-%d\n", issue.LineNumber, issue.ColumnStart, issue.ColumnEnd))
			summary.WriteString(fmt.Sprintf("   Type: %s\n", issue.Type))
			summary.WriteString(fmt.Sprintf("   Word: %s\n", issue.Word))
			summary.WriteString(fmt.Sprintf("   Context: %s\n", issue.Context))
			if len(issue.Suggestions) > 0 {
				summary.WriteString(fmt.Sprintf("   Suggestions: %s\n", strings.Join(issue.Suggestions, ", ")))
			}
			summary.WriteString("\n")
		}

		result.Content = append(result.Content, mcp.TextContent{
			Text: summary.String(),
			Type: "text",
		})
	} else {
		result.Content = append(result.Content, mcp.TextContent{
			Text: "No spelling issues found.",
			Type: "text",
		})
	}

	return result, nil
}

// RegisterSpellCheck registers the spellcheck tool with the MCP server
func RegisterSpellCheck(mcpServer *server.MCPServer) {
	// Create the tool definition
	spellCheckTool := mcp.NewTool("spellcheck",
		mcp.WithDescription("Checks spelling in code comments, string literals, and identifiers. Supports multiple programming languages and can detect misspellings in different naming conventions (camelCase, snake_case, PascalCase). Provides suggestions for corrections and can be customized with domain-specific dictionaries."),
		mcp.WithString("path",
			mcp.Description("The path of the file or directory to check (absolute or relative to working directory)"),
			mcp.Required(),
		),
		mcp.WithString("language",
			mcp.Description("The programming language to check (default: auto-detect from file extension)"),
		),
		mcp.WithBoolean("check_comments",
			mcp.Description("Whether to check spelling in comments (default: true)"),
		),
		mcp.WithBoolean("check_strings",
			mcp.Description("Whether to check spelling in string literals (default: true)"),
		),
		mcp.WithBoolean("check_identifiers",
			mcp.Description("Whether to check spelling in identifiers (variable and function names) (default: true)"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to check files recursively in subdirectories (default: true)"),
		),
		mcp.WithBoolean("use_relative_paths",
			mcp.Description("Whether to use relative paths in the results (default: true)"),
		),
		mcp.WithString("dictionary_type",
			mcp.Description("The type of dictionary to use: 'standard', 'programming', 'medical', etc. (default: 'standard')"),
		),
		mcp.WithArray("custom_dictionary",
			mcp.Description("A list of custom words to consider as correctly spelled"),
		),
		mcp.WithString("session_id",
			mcp.Description("Session ID to use for resolving relative paths"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("spellcheck", HandleSpellCheck)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(spellCheckTool, wrappedHandler)

	log.Printf("[SpellCheck] Registered spellcheck tool")
}
