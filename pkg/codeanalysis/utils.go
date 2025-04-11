package codeanalysis

import (
	"fmt"
	"strings"
)

// getLanguageFromExtension determines the programming language from a file extension
func getLanguageFromExtension(ext string) string {
	ext = strings.ToLower(ext)
	switch ext {
	case ".go":
		return "Go"
	case ".js":
		return "JavaScript"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".py":
		return "Python"
	case ".java":
		return "Java"
	case ".c", ".h":
		return "C"
	case ".cpp", ".hpp", ".cc":
		return "C++"
	case ".cs":
		return "C#"
	case ".php":
		return "PHP"
	case ".rb":
		return "Ruby"
	case ".swift":
		return "Swift"
	case ".kt", ".kts":
		return "Kotlin"
	case ".rs":
		return "Rust"
	case ".html", ".htm":
		return "HTML"
	case ".css":
		return "CSS"
	case ".json":
		return "JSON"
	case ".xml":
		return "XML"
	case ".yaml", ".yml":
		return "YAML"
	case ".md", ".markdown":
		return "Markdown"
	case ".sh":
		return "Shell"
	case ".bat", ".cmd":
		return "Batch"
	case ".ps1":
		return "PowerShell"
	default:
		return "Unknown"
	}
}

// isCodeFile checks if a file extension corresponds to a code file
func isCodeFile(ext string) bool {
	ext = strings.ToLower(ext)
	codeExtensions := map[string]bool{
		".go":    true,
		".js":    true,
		".ts":    true,
		".tsx":   true,
		".py":    true,
		".java":  true,
		".c":     true,
		".cpp":   true,
		".h":     true,
		".hpp":   true,
		".cc":    true,
		".cs":    true,
		".php":   true,
		".rb":    true,
		".swift": true,
		".kt":    true,
		".kts":   true,
		".rs":    true,
		".html":  true,
		".htm":   true,
		".css":   true,
		".json":  true,
		".xml":   true,
		".yaml":  true,
		".yml":   true,
		".md":    true,
		".sh":    true,
		".bat":   true,
		".cmd":   true,
		".ps1":   true,
	}

	return codeExtensions[ext]
}

// findComplexityIssues finds complexity issues in a file
func findComplexityIssues(filePath, content, language, severityLevel string) []IssueInfo {
	var issues []IssueInfo

	// Analyze file
	fileResult, err := analyzeFile(filePath)
	if err != nil {
		return issues
	}

	// Check overall complexity
	complexityThreshold := 5.0
	if severityLevel == "medium" {
		complexityThreshold = 3.0
	} else if severityLevel == "low" {
		complexityThreshold = 2.0
	}

	if fileResult.ComplexityScore > complexityThreshold {
		issues = append(issues, IssueInfo{
			Type:     "complexity",
			Message:  fmt.Sprintf("File has high complexity score (%.2f)", fileResult.ComplexityScore),
			FilePath: filePath,
			Line:     1,
			Snippet:  "",
		})
	}

	// Check function complexity
	for _, fn := range fileResult.TopFunctions {
		if fn.Complexity > complexityThreshold {
			issues = append(issues, IssueInfo{
				Type:     "complexity",
				Message:  fmt.Sprintf("Function '%s' has high complexity (%.2f)", fn.Name, fn.Complexity),
				FilePath: filePath,
				Line:     fn.Line,
				Snippet:  "",
			})
		}
	}

	return issues
}

// findDuplicationIssues finds code duplication issues in a file
func findDuplicationIssues(filePath, content, language, severityLevel string) []IssueInfo {
	var issues []IssueInfo

	// This is a simplified implementation
	// In a real implementation, you would use a more sophisticated algorithm
	// to detect code duplication

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Look for repeated blocks of code (simplified)
	blockSize := 5 // Look for blocks of 5 lines
	if len(lines) > blockSize*2 {
		for i := 0; i <= len(lines)-blockSize; i++ {
			block1 := strings.Join(lines[i:i+blockSize], "\n")

			for j := i + blockSize; j <= len(lines)-blockSize; j++ {
				block2 := strings.Join(lines[j:j+blockSize], "\n")

				// If blocks are similar
				if strings.TrimSpace(block1) == strings.TrimSpace(block2) {
					issues = append(issues, IssueInfo{
						Type:     "duplication",
						Message:  "Duplicated code block",
						FilePath: filePath,
						Line:     i + 1,
						Snippet:  block1,
					})

					// Only report one instance
					break
				}
			}
		}
	}

	return issues
}

// findNamingIssues finds naming convention issues in a file
func findNamingIssues(filePath, content, language, severityLevel string) []IssueInfo {
	var issues []IssueInfo

	// This is a simplified implementation
	// In a real implementation, you would use language-specific rules

	// Check for naming conventions based on language
	switch language {
	case "Go":
		// Check for non-camelCase or non-PascalCase identifiers
		nonCamelCaseRegex := `var\s+([a-z]+_[a-z]+)`
		matches := findRegexMatches(content, nonCamelCaseRegex)

		for _, match := range matches {
			issues = append(issues, IssueInfo{
				Type:     "naming",
				Message:  fmt.Sprintf("Variable '%s' does not follow Go naming conventions", match),
				FilePath: filePath,
				Line:     findLineNumber(content, match),
				Snippet:  "",
			})
		}
	case "JavaScript", "TypeScript":
		// Check for non-camelCase variables
		nonCamelCaseRegex := `var\s+([A-Z][a-z]*|[a-z]+_[a-z]+)`
		matches := findRegexMatches(content, nonCamelCaseRegex)

		for _, match := range matches {
			issues = append(issues, IssueInfo{
				Type:     "naming",
				Message:  fmt.Sprintf("Variable '%s' does not follow JavaScript naming conventions", match),
				FilePath: filePath,
				Line:     findLineNumber(content, match),
				Snippet:  "",
			})
		}
	}

	return issues
}

// findCommentIssues finds comment-related issues in a file
func findCommentIssues(filePath, content, language, severityLevel string) []IssueInfo {
	var issues []IssueInfo

	// Analyze file
	fileResult, err := analyzeFile(filePath)
	if err != nil {
		return issues
	}

	// Check comment ratio
	commentRatio := float64(fileResult.CommentLines) / float64(fileResult.LinesOfCode)

	// Low comment ratio
	if commentRatio < 0.1 {
		issues = append(issues, IssueInfo{
			Type:     "comments",
			Message:  fmt.Sprintf("Low comment ratio (%.2f%%)", commentRatio*100),
			FilePath: filePath,
			Line:     1,
			Snippet:  "",
		})
	}

	return issues
}

// findUnusedIssues finds unused code issues in a file
func findUnusedIssues(filePath, content, language, severityLevel string) []IssueInfo {
	var issues []IssueInfo

	// This is a simplified implementation
	// In a real implementation, you would use language-specific analysis

	// Check for TODO comments
	todoRegex := `(TODO|FIXME|XXX)`
	matches := findRegexMatches(content, todoRegex)

	for _, match := range matches {
		issues = append(issues, IssueInfo{
			Type:     "unused",
			Message:  fmt.Sprintf("Found '%s' comment", match),
			FilePath: filePath,
			Line:     findLineNumber(content, match),
			Snippet:  "",
		})
	}

	return issues
}

// suggestRefactoring suggests refactoring improvements for a file
func suggestRefactoring(filePath, content, language string) []SuggestionInfo {
	var suggestions []SuggestionInfo

	// Analyze file
	fileResult, err := analyzeFile(filePath)
	if err != nil {
		return suggestions
	}

	// Suggest refactoring for complex functions
	for _, fn := range fileResult.TopFunctions {
		if fn.Complexity > 3.0 {
			suggestions = append(suggestions, SuggestionInfo{
				Type:        "refactoring",
				Title:       fmt.Sprintf("Refactor complex function '%s'", fn.Name),
				Description: fmt.Sprintf("Function '%s' has a complexity score of %.2f. Consider breaking it down into smaller functions.", fn.Name, fn.Complexity),
				FilePath:    filePath,
				Line:        fn.Line,
				Before:      "",
				After:       "",
			})
		}
	}

	return suggestions
}

// suggestPerformanceImprovements suggests performance improvements for a file
func suggestPerformanceImprovements(filePath, content, language string) []SuggestionInfo {
	var suggestions []SuggestionInfo

	// This is a simplified implementation
	// In a real implementation, you would use language-specific analysis

	switch language {
	case "Go":
		// Check for inefficient string concatenation
		if strings.Contains(content, "+=") && strings.Contains(content, "string") {
			suggestions = append(suggestions, SuggestionInfo{
				Type:        "performance",
				Title:       "Use strings.Builder for string concatenation",
				Description: "Using += for string concatenation in loops is inefficient. Consider using strings.Builder instead.",
				FilePath:    filePath,
				Line:        findLineNumber(content, "+="),
				Before:      "s += \"text\"",
				After:       "var b strings.Builder\nb.WriteString(\"text\")\ns := b.String()",
			})
		}
	case "JavaScript", "TypeScript":
		// Check for inefficient array operations
		if strings.Contains(content, "Array(") && strings.Contains(content, "push") {
			suggestions = append(suggestions, SuggestionInfo{
				Type:        "performance",
				Title:       "Pre-allocate array size",
				Description: "Consider pre-allocating array size instead of using push in a loop.",
				FilePath:    filePath,
				Line:        findLineNumber(content, "Array("),
				Before:      "const arr = [];\nfor (let i = 0; i < 1000; i++) { arr.push(i); }",
				After:       "const arr = new Array(1000);\nfor (let i = 0; i < 1000; i++) { arr[i] = i; }",
			})
		}
	}

	return suggestions
}

// suggestReadabilityImprovements suggests readability improvements for a file
func suggestReadabilityImprovements(filePath, content, language string) []SuggestionInfo {
	var suggestions []SuggestionInfo

	// Check for long lines
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if len(line) > 100 {
			suggestions = append(suggestions, SuggestionInfo{
				Type:        "readability",
				Title:       "Long line",
				Description: fmt.Sprintf("Line %d is too long (%d characters). Consider breaking it into multiple lines.", i+1, len(line)),
				FilePath:    filePath,
				Line:        i + 1,
				Before:      line,
				After:       "// Break this line into multiple lines",
			})
		}
	}

	// Check for deep nesting
	if strings.Count(content, "{") > 10 {
		nestingLevel := 0
		maxNestingLevel := 0
		maxNestingLine := 0

		for i, line := range lines {
			nestingLevel += strings.Count(line, "{") - strings.Count(line, "}")
			if nestingLevel > maxNestingLevel {
				maxNestingLevel = nestingLevel
				maxNestingLine = i + 1
			}
		}

		if maxNestingLevel > 3 {
			suggestions = append(suggestions, SuggestionInfo{
				Type:        "readability",
				Title:       "Deep nesting",
				Description: fmt.Sprintf("Code has deep nesting (level %d). Consider refactoring to reduce nesting.", maxNestingLevel),
				FilePath:    filePath,
				Line:        maxNestingLine,
				Before:      "",
				After:       "",
			})
		}
	}

	return suggestions
}

// suggestMaintainabilityImprovements suggests maintainability improvements for a file
func suggestMaintainabilityImprovements(filePath, content, language string) []SuggestionInfo {
	var suggestions []SuggestionInfo

	// Check file size
	if len(content) > 1000 {
		suggestions = append(suggestions, SuggestionInfo{
			Type:        "maintainability",
			Title:       "Large file",
			Description: fmt.Sprintf("File is large (%d bytes). Consider splitting it into multiple files.", len(content)),
			FilePath:    filePath,
			Line:        1,
			Before:      "",
			After:       "",
		})
	}

	// Check for magic numbers
	magicNumberRegex := `\b\d{4,}\b`
	matches := findRegexMatches(content, magicNumberRegex)

	for _, match := range matches {
		suggestions = append(suggestions, SuggestionInfo{
			Type:        "maintainability",
			Title:       "Magic number",
			Description: fmt.Sprintf("Consider replacing magic number '%s' with a named constant.", match),
			FilePath:    filePath,
			Line:        findLineNumber(content, match),
			Before:      match,
			After:       "CONSTANT_NAME",
		})
	}

	return suggestions
}

// findRegexMatches finds all matches for a regex pattern in content
func findRegexMatches(content, pattern string) []string {
	var matches []string

	// This is a simplified implementation
	// In a real implementation, you would use the regexp package

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, pattern) {
			matches = append(matches, pattern)
		}
	}

	return matches
}

// findLineNumber finds the line number of a substring in content
func findLineNumber(content, substring string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, substring) {
			return i + 1
		}
	}
	return 1
}
