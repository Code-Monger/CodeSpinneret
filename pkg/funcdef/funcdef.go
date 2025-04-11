package funcdef

import (
	"context"
	"fmt"
	"io/ioutil"
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

// Language represents a programming language with its function definition patterns
type Language struct {
	Name                 string
	FileExtensions       []string
	FunctionStartPattern string
	FunctionEndPattern   string
	PrototypePattern     string // For languages like C/C++ that have separate prototypes
}

// GetSupportedLanguages returns a list of supported programming languages
func GetSupportedLanguages() []Language {
	return []Language{
		{
			Name:                 "Go",
			FileExtensions:       []string{".go"},
			FunctionStartPattern: `func\s+%s\s*\([^{]*{`,
			FunctionEndPattern:   `^}`,
		},
		{
			Name:                 "JavaScript",
			FileExtensions:       []string{".js", ".jsx", ".ts", ".tsx"},
			FunctionStartPattern: `(function\s+%s|const\s+%s\s*=\s*function|let\s+%s\s*=\s*function|var\s+%s\s*=\s*function|%s\s*=\s*function|%s\s*:\s*function|%s\s*\([^{]*\)\s*{|%s\s*=\s*\([^{]*\)\s*=>|%s\s*=>\s*{)`,
			FunctionEndPattern:   `^}`,
		},
		{
			Name:                 "Python",
			FileExtensions:       []string{".py"},
			FunctionStartPattern: `def\s+%s\s*\([^:]*:`,
			FunctionEndPattern:   `^(\S|$)`, // Non-indented line or empty line
		},
		{
			Name:                 "Java",
			FileExtensions:       []string{".java"},
			FunctionStartPattern: `(public|private|protected)?\s*(static)?\s*\S+\s+%s\s*\([^{]*{`,
			FunctionEndPattern:   `^[\t ]*}`,
		},
		{
			Name:                 "C#",
			FileExtensions:       []string{".cs"},
			FunctionStartPattern: `(public|private|protected|internal)?\s*(static|virtual|override|abstract)?\s*\S+\s+%s\s*\([^{]*{`,
			FunctionEndPattern:   `^[\t ]*}`,
		},
		{
			Name:                 "C/C++",
			FileExtensions:       []string{".c", ".cpp", ".cc", ".h", ".hpp"},
			FunctionStartPattern: `\S+\s+%s\s*\([^{;]*{`,
			FunctionEndPattern:   `^}`,
			PrototypePattern:     `\S+\s+%s\s*\([^{;]*;`,
		},
		{
			Name:                 "Ruby",
			FileExtensions:       []string{".rb"},
			FunctionStartPattern: `(def\s+%s|def\s+self\.%s)`,
			FunctionEndPattern:   `^end`,
		},
		{
			Name:                 "PHP",
			FileExtensions:       []string{".php"},
			FunctionStartPattern: `(function\s+%s|public\s+function\s+%s|private\s+function\s+%s|protected\s+function\s+%s|static\s+function\s+%s)\s*\([^{]*{`,
			FunctionEndPattern:   `^[\t ]*}`,
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

// FunctionDefinition represents a function definition
type FunctionDefinition struct {
	FilePath    string
	Language    string
	StartLine   int
	EndLine     int
	IsPrototype bool
	Content     string
}

// HandleFuncDef is the handler function for the funcdef tool
func HandleFuncDef(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract operation (get or replace)
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string (get or replace)")
	}
	if operation != "get" && operation != "replace" {
		return nil, fmt.Errorf("operation must be 'get' or 'replace'")
	}

	// Extract function name
	functionName, ok := arguments["function_name"].(string)
	if !ok {
		return nil, fmt.Errorf("function_name must be a string")
	}

	// Extract file path
	filePath, ok := arguments["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path must be a string")
	}

	// Extract language (optional)
	language := ""
	if langVal, ok := arguments["language"].(string); ok {
		language = langVal
	}

	// Extract include_prototype flag (for C/C++)
	includePrototype := false
	if protoVal, ok := arguments["include_prototype"].(bool); ok {
		includePrototype = protoVal
	}

	// Extract replacement content (for replace operation)
	replacementContent := ""
	if operation == "replace" {
		if replaceVal, ok := arguments["replacement_content"].(string); ok {
			replacementContent = replaceVal
		} else {
			return nil, fmt.Errorf("replacement_content must be provided for replace operation")
		}
	}

	// Get root directory from environment variable
	rootDir := os.Getenv(EnvRootDir)
	if rootDir == "" {
		rootDir = "." // Default to current directory if env var not set
	}

	// Resolve the file path
	var fullPath string
	if filepath.IsAbs(filePath) {
		fullPath = filePath
	} else {
		fullPath = filepath.Join(rootDir, filePath)
	}

	// Log the parameters for debugging
	log.Printf("[FuncDef] Operation: %s", operation)
	log.Printf("[FuncDef] Function name: %s", functionName)
	log.Printf("[FuncDef] File path: %s", fullPath)
	log.Printf("[FuncDef] Language: %s", language)
	log.Printf("[FuncDef] Include prototype: %v", includePrototype)

	// Determine the language if not specified
	if language == "" {
		ext := filepath.Ext(fullPath)
		if lang, found := GetLanguageByExtension(ext); found {
			language = lang.Name
			log.Printf("[FuncDef] Detected language: %s", language)
		} else {
			return nil, fmt.Errorf("could not determine language from file extension: %s", ext)
		}
	}

	// Get the language definition
	lang, found := GetLanguageByName(language)
	if !found {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Perform the operation
	if operation == "get" {
		// Get function definition
		definitions, err := GetFunctionDefinition(fullPath, functionName, lang, includePrototype)
		if err != nil {
			return nil, fmt.Errorf("error getting function definition: %v", err)
		}

		if len(definitions) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Function '%s' not found in file '%s'", functionName, filePath),
					},
				},
			}, nil
		}

		// Format the result
		resultText := fmt.Sprintf("Function '%s' in file '%s':\n\n", functionName, filePath)
		for i, def := range definitions {
			if len(definitions) > 1 {
				if def.IsPrototype {
					resultText += fmt.Sprintf("Prototype (lines %d-%d):\n", def.StartLine, def.EndLine)
				} else {
					resultText += fmt.Sprintf("Implementation (lines %d-%d):\n", def.StartLine, def.EndLine)
				}
			} else {
				resultText += fmt.Sprintf("Lines %d-%d:\n", def.StartLine, def.EndLine)
			}
			resultText += def.Content
			if i < len(definitions)-1 {
				resultText += "\n\n"
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
	} else {
		// First, get the original function definition before replacement
		originalDefs, err := GetFunctionDefinition(fullPath, functionName, lang, includePrototype)
		if err != nil {
			return nil, fmt.Errorf("error getting original function definition: %v", err)
		}

		if len(originalDefs) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Function '%s' not found in file '%s'", functionName, filePath),
					},
				},
			}, nil
		}

		// Store original content
		originalContent := make(map[bool]string) // map[isPrototype]content
		for _, def := range originalDefs {
			originalContent[def.IsPrototype] = def.Content
		}

		// Perform the replacement
		success, err := ReplaceFunctionDefinition(fullPath, functionName, lang, replacementContent, includePrototype)
		if err != nil {
			return nil, fmt.Errorf("error replacing function definition: %v", err)
		}

		if !success {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Function '%s' not found in file '%s'", functionName, filePath),
					},
				},
			}, nil
		}

		// Get the updated function definition to confirm the change
		updatedDefs, err := GetFunctionDefinition(fullPath, functionName, lang, includePrototype)
		if err != nil {
			return nil, fmt.Errorf("error getting updated function definition: %v", err)
		}

		// Format the result with before/after comparison
		resultText := fmt.Sprintf("Function '%s' in file '%s' has been replaced.\n\n", functionName, filePath)

		for _, def := range updatedDefs {
			if def.IsPrototype {
				resultText += "Prototype "
			} else {
				resultText += "Implementation "
			}
			resultText += fmt.Sprintf("(lines %d-%d):\n\n", def.StartLine, def.EndLine)

			// Show before/after comparison
			resultText += "BEFORE:\n"
			resultText += "```\n"
			if orig, ok := originalContent[def.IsPrototype]; ok {
				resultText += orig
			} else {
				resultText += "[Not found]"
			}
			resultText += "\n```\n\n"

			resultText += "AFTER:\n"
			resultText += "```\n"
			resultText += def.Content
			resultText += "\n```\n\n"
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
}

// GetFunctionDefinition gets the definition of a function from a file
func GetFunctionDefinition(filePath, functionName string, language Language, includePrototype bool) ([]FunctionDefinition, error) {
	var definitions []FunctionDefinition

	// Read the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Split the content into lines
	lines := strings.Split(string(content), "\n")

	// Compile the function start pattern
	startPatternStr := fmt.Sprintf(language.FunctionStartPattern, regexp.QuoteMeta(functionName))
	for i := 1; i < 10; i++ {
		// Handle multiple placeholders in the pattern (for JavaScript)
		startPatternStr = strings.Replace(startPatternStr, fmt.Sprintf("%%s", i), regexp.QuoteMeta(functionName), -1)
	}
	startPattern, err := regexp.Compile(startPatternStr)
	if err != nil {
		return nil, fmt.Errorf("error compiling function start pattern: %v", err)
	}

	// Compile the function end pattern
	endPattern, err := regexp.Compile(language.FunctionEndPattern)
	if err != nil {
		return nil, fmt.Errorf("error compiling function end pattern: %v", err)
	}

	// Compile the prototype pattern if available and requested
	var prototypePattern *regexp.Regexp
	if includePrototype && language.PrototypePattern != "" {
		protoPatternStr := fmt.Sprintf(language.PrototypePattern, regexp.QuoteMeta(functionName))
		prototypePattern, err = regexp.Compile(protoPatternStr)
		if err != nil {
			return nil, fmt.Errorf("error compiling prototype pattern: %v", err)
		}
	}

	// Find function definitions
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for prototype
		if prototypePattern != nil && prototypePattern.MatchString(line) {
			def := FunctionDefinition{
				FilePath:    filePath,
				Language:    language.Name,
				StartLine:   i + 1,
				EndLine:     i + 1,
				IsPrototype: true,
				Content:     line,
			}
			definitions = append(definitions, def)
			continue
		}

		// Check for function start
		if startPattern.MatchString(line) {
			startLine := i
			endLine := i
			indentLevel := 0
			braceCount := 0

			// Count opening braces in the start line
			for _, char := range line {
				if char == '{' {
					braceCount++
				}
			}

			// Find the end of the function
			inMultiLineComment := false
			inString := false
			inRawString := false
			inChar := false
			stringDelimiter := byte(0)

			for j := i + 1; j < len(lines); j++ {
				currentLine := lines[j]
				trimmedLine := strings.TrimSpace(currentLine)

				// Skip empty lines
				if trimmedLine == "" {
					continue
				}

				// Handle comments and strings for languages that use // and /* */ style comments
				if language.Name == "Go" || language.Name == "C/C++" || language.Name == "Java" ||
					language.Name == "C#" || language.Name == "JavaScript" || language.Name == "PHP" {

					// Process the line character by character to handle comments, strings, and braces
					inLineComment := false
					inBlockComment := inMultiLineComment
					lineCharCount := 0
					escaped := false

					for k := 0; k < len(currentLine); k++ {
						// Handle escape sequences
						if escaped {
							escaped = false
							continue
						}

						// Check for escape character
						if (inString || inChar || inRawString) && currentLine[k] == '\\' && !inRawString {
							escaped = true
							continue
						}

						// Skip processing if in a line comment
						if inLineComment {
							continue
						}

						// Check for start of single-line comment (if not in a string)
						if !inString && !inChar && !inRawString && k < len(currentLine)-1 &&
							currentLine[k] == '/' && currentLine[k+1] == '/' && !inBlockComment {
							inLineComment = true
							continue
						}

						// Check for start of multi-line comment (if not in a string)
						if !inString && !inChar && !inRawString && k < len(currentLine)-1 &&
							currentLine[k] == '/' && currentLine[k+1] == '*' && !inBlockComment {
							inBlockComment = true
							inMultiLineComment = true
							k++ // Skip the '*'
							continue
						}

						// Check for end of multi-line comment
						if inBlockComment && k < len(currentLine)-1 &&
							currentLine[k] == '*' && currentLine[k+1] == '/' {
							inBlockComment = false
							inMultiLineComment = false
							k++ // Skip the '/'
							continue
						}

						// Handle string literals
						if !inBlockComment && !inLineComment {
							// Check for start/end of string literals
							if !inString && !inChar && !inRawString {
								// Start of double-quoted string
								if currentLine[k] == '"' {
									inString = true
									stringDelimiter = '"'
									continue
								}

								// Start of single-quoted character/string
								if currentLine[k] == '\'' {
									inChar = true
									stringDelimiter = '\''
									continue
								}

								// Start of raw string (Go backtick or JS template literal)
								if language.Name == "Go" && currentLine[k] == '`' {
									inRawString = true
									stringDelimiter = '`'
									continue
								}
							} else {
								// End of string/char/raw string
								if currentLine[k] == stringDelimiter {
									if inString {
										inString = false
									} else if inChar {
										inChar = false
									} else if inRawString {
										inRawString = false
									}
									stringDelimiter = 0
									continue
								}
							}
						}

						// Only count braces if not in a comment or string
						if !inBlockComment && !inLineComment && !inString && !inChar && !inRawString {
							if currentLine[k] == '{' {
								braceCount++
								lineCharCount++
							} else if currentLine[k] == '}' {
								braceCount--
								lineCharCount++
							} else {
								lineCharCount++
							}
						}
					}

					// For languages with explicit end markers, check if we've reached the end
					if !inMultiLineComment && !inLineComment && endPattern.MatchString(currentLine) && lineCharCount > 0 {
						if braceCount <= 0 {
							endLine = j
							break
						}
					}
				} else if language.Name == "Python" {
					// Skip comment lines
					if strings.HasPrefix(trimmedLine, "#") {
						continue
					}

					if startLine == i { // First line after function definition
						if len(currentLine) > 0 && !strings.HasPrefix(currentLine, " ") && !strings.HasPrefix(currentLine, "\t") {
							// Non-indented line means end of function
							break
						}
						if len(currentLine) > 0 {
							// Determine the indentation level of the function body
							indentLevel = countLeadingWhitespace(currentLine)
						}
					} else {
						if len(currentLine) > 0 && countLeadingWhitespace(currentLine) <= indentLevel {
							// Line with less indentation means end of function
							break
						}
					}
				} else if language.Name == "Ruby" {
					// Skip comment lines
					if strings.HasPrefix(trimmedLine, "#") {
						continue
					}

					if endPattern.MatchString(currentLine) {
						endLine = j
						break
					}
				} else if endPattern.MatchString(currentLine) {
					endLine = j
					break
				}

				endLine = j
			}

			// Extract the function content
			functionContent := strings.Join(lines[startLine:endLine+1], "\n")

			def := FunctionDefinition{
				FilePath:    filePath,
				Language:    language.Name,
				StartLine:   startLine + 1,
				EndLine:     endLine + 1,
				IsPrototype: false,
				Content:     functionContent,
			}
			definitions = append(definitions, def)

			// Skip to the end of the function
			i = endLine
		}
	}

	return definitions, nil
}

// ReplaceFunctionDefinition replaces the definition of a function in a file
func ReplaceFunctionDefinition(filePath, functionName string, language Language, replacementContent string, includePrototype bool) (bool, error) {
	// Get the current function definition
	definitions, err := GetFunctionDefinition(filePath, functionName, language, includePrototype)
	if err != nil {
		return false, fmt.Errorf("error getting function definition: %v", err)
	}

	if len(definitions) == 0 {
		return false, nil
	}

	// Read the file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file: %v", err)
	}

	// Split the content into lines
	lines := strings.Split(string(content), "\n")

	// Sort definitions in reverse order to avoid line number changes
	// when replacing multiple definitions
	for i := len(definitions) - 1; i >= 0; i-- {
		def := definitions[i]

		// Skip prototypes if not requested
		if def.IsPrototype && !includePrototype {
			continue
		}

		// Replace the function definition
		if def.IsPrototype {
			// For prototypes, replace just the one line
			lines[def.StartLine-1] = replacementContent
		} else {
			// For implementations, replace the entire function
			newLines := strings.Split(replacementContent, "\n")
			lines = append(lines[:def.StartLine-1], append(newLines, lines[def.EndLine:]...)...)
		}
	}

	// Write the modified content back to the file
	err = ioutil.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		return false, fmt.Errorf("error writing file: %v", err)
	}

	return true, nil
}

// countLeadingWhitespace counts the number of leading whitespace characters in a string
func countLeadingWhitespace(s string) int {
	count := 0
	for _, char := range s {
		if char == ' ' {
			count++
		} else if char == '\t' {
			count += 4 // Count tabs as 4 spaces
		} else {
			break
		}
	}
	return count
}

// RegisterFuncDef registers the funcdef tool with the MCP server
func RegisterFuncDef(mcpServer *server.MCPServer) {
	// Create the tool definition
	funcDefTool := mcp.NewTool("funcdef",
		mcp.WithDescription("Gets or replaces function definitions in source code files"),
		mcp.WithString("operation",
			mcp.Description("The operation to perform: 'get' or 'replace'"),
			mcp.Required(),
		),
		mcp.WithString("function_name",
			mcp.Description("The name of the function to get or replace"),
			mcp.Required(),
		),
		mcp.WithString("file_path",
			mcp.Description("The path of the file containing the function"),
			mcp.Required(),
		),
		mcp.WithString("language",
			mcp.Description("The programming language (if not specified, will be determined from file extension)"),
		),
		mcp.WithBoolean("include_prototype",
			mcp.Description("Whether to include function prototypes (for C/C++)"),
		),
		mcp.WithString("replacement_content",
			mcp.Description("The new content to replace the function with (for replace operation)"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("funcdef", HandleFuncDef)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(funcDefTool, wrappedHandler)

	log.Printf("[FuncDef] Registered funcdef tool")
}
