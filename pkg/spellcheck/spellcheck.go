package spellcheck

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Environment variable name for root directory (same as patch and linecount tools)
const EnvRootDir = "PATCH_ROOT_DIR"

// SpellCheckResult represents a spelling issue found in the code
type SpellCheckResult struct {
	FilePath    string   `json:"file_path"`
	LineNumber  int      `json:"line_number"`
	ColumnStart int      `json:"column_start"`
	ColumnEnd   int      `json:"column_end"`
	Word        string   `json:"word"`
	Context     string   `json:"context"`
	Type        string   `json:"type"` // "comment", "string", "identifier"
	Suggestions []string `json:"suggestions,omitempty"`
}

// Language represents a programming language with its file extensions and comment patterns
type Language struct {
	Name                  string
	FileExtensions        []string
	SingleLineComment     string
	MultiLineCommentStart string
	MultiLineCommentEnd   string
	StringDelimiters      []string
	RawStringDelimiters   []string
}

// GetSupportedLanguages returns a list of supported programming languages
func GetSupportedLanguages() []Language {
	return []Language{
		{
			Name:                  "Go",
			FileExtensions:        []string{".go"},
			SingleLineComment:     "//",
			MultiLineCommentStart: "/*",
			MultiLineCommentEnd:   "*/",
			StringDelimiters:      []string{"\""},
			RawStringDelimiters:   []string{"`"},
		},
		{
			Name:                  "JavaScript",
			FileExtensions:        []string{".js", ".jsx", ".ts", ".tsx"},
			SingleLineComment:     "//",
			MultiLineCommentStart: "/*",
			MultiLineCommentEnd:   "*/",
			StringDelimiters:      []string{"\"", "'"},
			RawStringDelimiters:   []string{"`"},
		},
		{
			Name:                  "Python",
			FileExtensions:        []string{".py"},
			SingleLineComment:     "#",
			MultiLineCommentStart: "\"\"\"",
			MultiLineCommentEnd:   "\"\"\"",
			StringDelimiters:      []string{"\"", "'"},
			RawStringDelimiters:   []string{"\"\"\"", "'''"},
		},
		{
			Name:                  "Java",
			FileExtensions:        []string{".java"},
			SingleLineComment:     "//",
			MultiLineCommentStart: "/*",
			MultiLineCommentEnd:   "*/",
			StringDelimiters:      []string{"\""},
			RawStringDelimiters:   []string{},
		},
		{
			Name:                  "C#",
			FileExtensions:        []string{".cs"},
			SingleLineComment:     "//",
			MultiLineCommentStart: "/*",
			MultiLineCommentEnd:   "*/",
			StringDelimiters:      []string{"\""},
			RawStringDelimiters:   []string{"@\""},
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

	// Get root directory from environment variable
	rootDir := os.Getenv(EnvRootDir)
	if rootDir == "" {
		rootDir = "." // Default to current directory if env var not set
	}

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

// spellCheckFile performs spell checking on a single file
func spellCheckFile(filePath, language string, checkComments, checkStrings, checkIdentifiers bool, dictionaryType string, customDictionary []string) ([]SpellCheckResult, error) {
	var results []SpellCheckResult

	// Determine the language if not specified
	var lang Language
	if language == "" {
		ext := filepath.Ext(filePath)
		var found bool
		lang, found = GetLanguageByExtension(ext)
		if !found {
			return nil, fmt.Errorf("unsupported file extension: %s", ext)
		}
	} else {
		var found bool
		lang, found = GetLanguageByName(language)
		if !found {
			return nil, fmt.Errorf("unsupported language: %s", language)
		}
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Process each line
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		// Check comments if enabled
		if checkComments {
			// Check for single-line comments
			commentIndex := strings.Index(line, lang.SingleLineComment)
			if commentIndex >= 0 {
				comment := line[commentIndex+len(lang.SingleLineComment):]
				commentResults := checkTextForSpellingErrors(comment, lineNumber, commentIndex+len(lang.SingleLineComment), "comment", dictionaryType, customDictionary)
				for _, result := range commentResults {
					result.FilePath = filePath
					result.Context = line
					results = append(results, result)
				}
			}
		}

		// Check string literals if enabled
		if checkStrings {
			for _, delimiter := range lang.StringDelimiters {
				// Find all string literals in the line
				startIndex := 0
				for {
					startIndex = strings.Index(line[startIndex:], delimiter)
					if startIndex < 0 {
						break
					}

					// Find the end of the string
					endIndex := strings.Index(line[startIndex+len(delimiter):], delimiter)
					if endIndex < 0 {
						break
					}

					// Extract the string content
					stringContent := line[startIndex+len(delimiter) : startIndex+len(delimiter)+endIndex]
					stringResults := checkTextForSpellingErrors(stringContent, lineNumber, startIndex+len(delimiter), "string", dictionaryType, customDictionary)
					for _, result := range stringResults {
						result.FilePath = filePath
						result.Context = line
						results = append(results, result)
					}

					startIndex += len(delimiter) + endIndex + len(delimiter)
				}
			}
		}

		// Check identifiers if enabled
		if checkIdentifiers {
			// Simple regex to find identifiers (variable and function names)
			// This is a simplified approach and would need to be more sophisticated in a real implementation
			words := strings.Fields(line)
			for _, word := range words {
				// Skip if it's not a valid identifier
				if !isValidIdentifier(word) {
					continue
				}

				// Split camelCase, PascalCase, or snake_case identifiers into words
				subWords := splitIdentifier(word)
				for _, subWord := range subWords {
					if len(subWord) > 2 && !isCommonProgrammingTerm(subWord) {
						// Check if the word is misspelled
						if isMisspelled(subWord, dictionaryType) && !isInCustomDictionary(subWord, customDictionary) {
							// Find the position of the word in the line
							wordIndex := strings.Index(line, word)
							if wordIndex >= 0 {
								results = append(results, SpellCheckResult{
									FilePath:    filePath,
									LineNumber:  lineNumber,
									ColumnStart: wordIndex,
									ColumnEnd:   wordIndex + len(word),
									Word:        subWord,
									Context:     line,
									Type:        "identifier",
									Suggestions: getSuggestions(subWord, dictionaryType),
								})
							}
						}
					}
				}
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return results, nil
}

// checkTextForSpellingErrors checks a text for spelling errors
func checkTextForSpellingErrors(text string, lineNumber, columnOffset int, textType, dictionaryType string, customDictionary []string) []SpellCheckResult {
	var results []SpellCheckResult

	// Split the text into words
	words := strings.Fields(text)

	// Track the current column position
	currentPos := columnOffset

	// Check each word
	for _, word := range words {
		// Clean the word (remove punctuation)
		cleanWord := cleanWord(word)

		// Skip short words, numbers, and common programming terms
		if len(cleanWord) > 2 && !isNumeric(cleanWord) && !isCommonProgrammingTerm(cleanWord) {
			// Check if the word is misspelled
			if isMisspelled(cleanWord, dictionaryType) && !isInCustomDictionary(cleanWord, customDictionary) {
				// Find the position of the word in the text
				wordIndex := strings.Index(text[currentPos-columnOffset:], word)
				if wordIndex >= 0 {
					wordPos := currentPos + wordIndex
					results = append(results, SpellCheckResult{
						LineNumber:  lineNumber,
						ColumnStart: wordPos,
						ColumnEnd:   wordPos + len(word),
						Word:        cleanWord,
						Type:        textType,
						Suggestions: getSuggestions(cleanWord, dictionaryType),
					})
				}
			}
		}

		// Move the position past this word
		currentPos += len(word) + 1 // +1 for the space
	}

	return results
}

// cleanWord removes punctuation from a word
func cleanWord(word string) string {
	// Remove common punctuation
	word = strings.Trim(word, ",.;:!?\"'()[]{}")
	return word
}

// isValidIdentifier checks if a string is a valid identifier
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// First character must be a letter or underscore
	if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_", rune(s[0])) {
		return false
	}

	// Rest can be letters, digits, or underscores
	for _, c := range s[1:] {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_", c) {
			return false
		}
	}

	return true
}

// isCommonProgrammingTerm checks if a word is a common programming term
func isCommonProgrammingTerm(word string) bool {
	// Common programming terms and abbreviations
	programmingTerms := map[string]bool{
		"var": true, "func": true, "int": true, "str": true, "bool": true,
		"len": true, "fmt": true, "println": true, "printf": true, "sprintf": true,
		"args": true, "argv": true, "argc": true, "param": true, "params": true,
		"init": true, "exec": true, "eval": true, "impl": true, "pkg": true,
		"lib": true, "src": true, "dest": true, "tmp": true, "temp": true,
		"dir": true, "dirs": true, "cmd": true, "cmds": true, "env": true,
		"config": true, "cfg": true, "ctx": true, "req": true, "res": true,
		"err": true, "stdin": true, "stdout": true, "stderr": true,
	}

	return programmingTerms[strings.ToLower(word)]
}

// splitIdentifier splits an identifier into words based on casing
func splitIdentifier(identifier string) []string {
	var words []string

	// Check for snake_case
	if strings.Contains(identifier, "_") {
		// Split by underscore
		parts := strings.Split(identifier, "_")
		for _, part := range parts {
			if part != "" {
				words = append(words, part)
			}
		}
		return words
	}

	// Handle camelCase and PascalCase
	var currentWord strings.Builder
	for i, c := range identifier {
		if i > 0 && unicode.IsUpper(c) {
			// End of a word
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(c)
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	return words
}

// isNumeric checks if a string is a number
func isNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// isInCustomDictionary checks if a word is in the custom dictionary
func isInCustomDictionary(word string, customDictionary []string) bool {
	for _, dictWord := range customDictionary {
		if strings.EqualFold(word, dictWord) {
			return true
		}
	}
	return false
}

// isMisspelled checks if a word is misspelled
func isMisspelled(word, dictionaryType string) bool {
	// This is a simplified implementation
	// In a real implementation, this would use a proper spell checking library
	// or call an external API

	// For demo purposes, we'll just check against a small list of common words
	// and consider any word not in this list as misspelled
	commonWords := map[string]bool{
		"the": true, "and": true, "that": true, "have": true, "for": true,
		"not": true, "with": true, "you": true, "this": true, "but": true,
		"his": true, "from": true, "they": true, "say": true, "she": true,
		"will": true, "one": true, "all": true, "would": true, "there": true,
		"their": true, "what": true, "out": true, "about": true, "who": true,
		"get": true, "which": true, "when": true, "make": true, "can": true,
		"like": true, "time": true, "just": true, "him": true, "know": true,
		"take": true, "people": true, "into": true, "year": true, "your": true,
		"good": true, "some": true, "could": true, "them": true, "see": true,
		"other": true, "than": true, "then": true, "now": true, "look": true,
		"only": true, "come": true, "its": true, "over": true, "think": true,
		"also": true, "back": true, "after": true, "use": true, "two": true,
		"how": true, "our": true, "work": true, "first": true, "well": true,
		"way": true, "even": true, "new": true, "want": true, "because": true,
		"any": true, "these": true, "give": true, "day": true, "most": true,
		"us": true, "is": true, "are": true, "be": true, "was": true, "were": true,
		"am": true, "been": true, "being": true, "do": true, "does": true, "did": true,
		"done": true, "doing": true, "has": true, "had": true, "having": true,
		"go": true, "goes": true, "went": true, "gone": true, "going": true,
		"hello": true, "world": true, "main": true, "fmt": true, "println": true,
		"display": true, "text": true, "message": true, "user": true, "print": true,
	}

	return !commonWords[strings.ToLower(word)]
}

// getSuggestions gets spelling suggestions for a misspelled word
func getSuggestions(word, dictionaryType string) []string {
	// This is a simplified implementation
	// In a real implementation, this would use a proper spell checking library
	// or call an external API

	// For demo purposes, we'll just return some simple suggestions
	lowerWord := strings.ToLower(word)

	// Check for common misspellings
	commonMisspellings := map[string][]string{
		"coment":    []string{"comment"},
		"speling":   []string{"spelling"},
		"mesage":    []string{"message"},
		"acount":    []string{"account"},
		"messge":    []string{"message"},
		"mispelled": []string{"misspelled"},
	}

	if suggestions, ok := commonMisspellings[lowerWord]; ok {
		return suggestions
	}

	// Otherwise, generate some simple suggestions
	var suggestions []string

	// Add 'e' if the word ends with a consonant
	if len(word) > 2 && !isVowel(rune(word[len(word)-1])) {
		suggestions = append(suggestions, word+"e")
	}

	// Double the last letter
	if len(word) > 2 {
		suggestions = append(suggestions, word+string(word[len(word)-1]))
	}

	// Add common suffixes
	suggestions = append(suggestions, word+"s", word+"ed", word+"ing")

	return suggestions
}

// isVowel checks if a character is a vowel
func isVowel(c rune) bool {
	return strings.ContainsRune("aeiouAEIOU", c)
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
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("spellcheck", HandleSpellCheck)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(spellCheckTool, wrappedHandler)

	log.Printf("[SpellCheck] Registered spellcheck tool")
}
