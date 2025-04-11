package codeanalysis

import (
	"context"
	"fmt"

	"github.com/Code-Monger/CodeSpinneret/pkg/stats"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// HandleCodeAnalysis is the handler function for the code analysis tool
func HandleCodeAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments := request.Params.Arguments

	// Extract operation
	operation, ok := arguments["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	switch operation {
	case "analyze_file":
		return handleAnalyzeFile(arguments)
	case "analyze_directory":
		return handleAnalyzeDirectory(arguments)
	case "find_issues":
		return handleFindIssues(arguments)
	case "suggest_improvements":
		return handleSuggestImprovements(arguments)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

// handleAnalyzeFile handles the analyze_file operation
func handleAnalyzeFile(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract file path
	filePath, ok := arguments["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path must be a string")
	}

	// Analyze the file
	analysisResult, err := analyzeFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error analyzing file: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Code Analysis Results for: %s\n\n", filePath)
	resultText += fmt.Sprintf("Language: %s\n", analysisResult.Language)
	resultText += fmt.Sprintf("Lines of code: %d\n", analysisResult.LinesOfCode)
	resultText += fmt.Sprintf("Functions/methods: %d\n", analysisResult.FunctionCount)
	resultText += fmt.Sprintf("Classes/structs: %d\n", analysisResult.ClassCount)
	resultText += fmt.Sprintf("Comments: %d lines\n", analysisResult.CommentLines)
	resultText += fmt.Sprintf("Complexity score: %.2f\n", analysisResult.ComplexityScore)
	resultText += fmt.Sprintf("Time taken: %s\n\n", analysisResult.TimeTaken)

	if len(analysisResult.TopFunctions) > 0 {
		resultText += "Top complex functions:\n"
		for _, function := range analysisResult.TopFunctions {
			resultText += fmt.Sprintf("- %s (complexity: %.2f)\n", function.Name, function.Complexity)
			resultText += fmt.Sprintf("  Line: %d\n", function.Line)
		}
		resultText += "\n"
	}

	if len(analysisResult.Dependencies) > 0 {
		resultText += "Dependencies:\n"
		for _, dep := range analysisResult.Dependencies {
			resultText += fmt.Sprintf("- %s\n", dep)
		}
		resultText += "\n"
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

// handleAnalyzeDirectory handles the analyze_directory operation
func handleAnalyzeDirectory(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract directory path
	dirPath, ok := arguments["directory_path"].(string)
	if !ok {
		return nil, fmt.Errorf("directory_path must be a string")
	}

	// Extract file patterns
	var filePatterns []string
	if patterns, ok := arguments["file_patterns"].([]interface{}); ok {
		for _, pattern := range patterns {
			if patternStr, ok := pattern.(string); ok {
				filePatterns = append(filePatterns, patternStr)
			}
		}
	}
	if len(filePatterns) == 0 {
		// Default to common code file patterns
		filePatterns = []string{"*.go", "*.js", "*.ts", "*.py", "*.java", "*.c", "*.cpp", "*.h", "*.cs"}
	}

	// Extract recursive flag
	recursive := true // Default to recursive
	if recursiveBool, ok := arguments["recursive"].(bool); ok {
		recursive = recursiveBool
	}

	// Analyze the directory
	analysisResult, err := analyzeDirectory(dirPath, filePatterns, recursive)
	if err != nil {
		return nil, fmt.Errorf("error analyzing directory: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Code Analysis Results for Directory: %s\n\n", dirPath)
	resultText += fmt.Sprintf("Files analyzed: %d\n", analysisResult.FilesAnalyzed)
	resultText += fmt.Sprintf("Total lines of code: %d\n", analysisResult.TotalLinesOfCode)
	resultText += fmt.Sprintf("Total functions/methods: %d\n", analysisResult.TotalFunctionCount)
	resultText += fmt.Sprintf("Total classes/structs: %d\n", analysisResult.TotalClassCount)
	resultText += fmt.Sprintf("Total comments: %d lines\n", analysisResult.TotalCommentLines)
	resultText += fmt.Sprintf("Average complexity score: %.2f\n", analysisResult.AverageComplexity)
	resultText += fmt.Sprintf("Time taken: %s\n\n", analysisResult.TimeTaken)

	if len(analysisResult.LanguageBreakdown) > 0 {
		resultText += "Language breakdown:\n"
		for lang, count := range analysisResult.LanguageBreakdown {
			resultText += fmt.Sprintf("- %s: %d files\n", lang, count)
		}
		resultText += "\n"
	}

	if len(analysisResult.TopComplexFiles) > 0 {
		resultText += "Top complex files:\n"
		for _, file := range analysisResult.TopComplexFiles {
			resultText += fmt.Sprintf("- %s (complexity: %.2f)\n", file.Path, file.Complexity)
			resultText += fmt.Sprintf("  Lines of code: %d\n", file.LinesOfCode)
		}
		resultText += "\n"
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

// handleFindIssues handles the find_issues operation
func handleFindIssues(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract target path (file or directory)
	targetPath, ok := arguments["target_path"].(string)
	if !ok {
		return nil, fmt.Errorf("target_path must be a string")
	}

	// Extract issue types
	var issueTypes []string
	if types, ok := arguments["issue_types"].([]interface{}); ok {
		for _, t := range types {
			if typeStr, ok := t.(string); ok {
				issueTypes = append(issueTypes, typeStr)
			}
		}
	}
	if len(issueTypes) == 0 {
		// Default to all issue types
		issueTypes = []string{"complexity", "duplication", "naming", "comments", "unused"}
	}

	// Extract severity level
	severityLevel := "medium" // Default to medium severity
	if severity, ok := arguments["severity"].(string); ok {
		severityLevel = severity
	}

	// Find issues
	issuesResult, err := findIssues(targetPath, issueTypes, severityLevel)
	if err != nil {
		return nil, fmt.Errorf("error finding issues: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Code Issues Found in: %s\n\n", targetPath)
	resultText += fmt.Sprintf("Total issues: %d\n", issuesResult.TotalIssues)
	resultText += fmt.Sprintf("Severity level: %s\n", severityLevel)
	resultText += fmt.Sprintf("Time taken: %s\n\n", issuesResult.TimeTaken)

	if len(issuesResult.IssuesByType) > 0 {
		resultText += "Issues by type:\n"
		for issueType, count := range issuesResult.IssuesByType {
			resultText += fmt.Sprintf("- %s: %d issues\n", issueType, count)
		}
		resultText += "\n"
	}

	if len(issuesResult.Issues) > 0 {
		resultText += "Issues:\n"
		for i, issue := range issuesResult.Issues {
			resultText += fmt.Sprintf("%d. [%s] %s\n", i+1, issue.Type, issue.Message)
			resultText += fmt.Sprintf("   File: %s, Line: %d\n", issue.FilePath, issue.Line)
			if issue.Snippet != "" {
				resultText += fmt.Sprintf("   Code: %s\n", issue.Snippet)
			}
			resultText += "\n"
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

// handleSuggestImprovements handles the suggest_improvements operation
func handleSuggestImprovements(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	// Extract target path (file or directory)
	targetPath, ok := arguments["target_path"].(string)
	if !ok {
		return nil, fmt.Errorf("target_path must be a string")
	}

	// Extract improvement types
	var improvementTypes []string
	if types, ok := arguments["improvement_types"].([]interface{}); ok {
		for _, t := range types {
			if typeStr, ok := t.(string); ok {
				improvementTypes = append(improvementTypes, typeStr)
			}
		}
	}
	if len(improvementTypes) == 0 {
		// Default to all improvement types
		improvementTypes = []string{"refactoring", "performance", "readability", "maintainability"}
	}

	// Suggest improvements
	improvementsResult, err := suggestImprovements(targetPath, improvementTypes)
	if err != nil {
		return nil, fmt.Errorf("error suggesting improvements: %v", err)
	}

	// Format the result
	resultText := fmt.Sprintf("Suggested Code Improvements for: %s\n\n", targetPath)
	resultText += fmt.Sprintf("Total suggestions: %d\n", improvementsResult.TotalSuggestions)
	resultText += fmt.Sprintf("Time taken: %s\n\n", improvementsResult.TimeTaken)

	if len(improvementsResult.SuggestionsByType) > 0 {
		resultText += "Suggestions by type:\n"
		for improvementType, count := range improvementsResult.SuggestionsByType {
			resultText += fmt.Sprintf("- %s: %d suggestions\n", improvementType, count)
		}
		resultText += "\n"
	}

	if len(improvementsResult.Suggestions) > 0 {
		resultText += "Suggestions:\n"
		for i, suggestion := range improvementsResult.Suggestions {
			resultText += fmt.Sprintf("%d. [%s] %s\n", i+1, suggestion.Type, suggestion.Title)
			resultText += fmt.Sprintf("   File: %s, Line: %d\n", suggestion.FilePath, suggestion.Line)
			resultText += fmt.Sprintf("   Description: %s\n", suggestion.Description)
			if suggestion.Before != "" && suggestion.After != "" {
				resultText += "   Before:\n```\n" + suggestion.Before + "\n```\n"
				resultText += "   After:\n```\n" + suggestion.After + "\n```\n"
			}
			resultText += "\n"
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

// RegisterCodeAnalysis registers the code analysis tool with the MCP server
func RegisterCodeAnalysis(mcpServer *server.MCPServer) {
	// Create the tool definition
	codeAnalysisTool := mcp.NewTool("codeanalysis",
		mcp.WithDescription("Analyzes code to provide insights, metrics, and suggestions for improvement"),
		mcp.WithString("operation",
			mcp.Description("Operation to perform: 'analyze_file', 'analyze_directory', 'find_issues', or 'suggest_improvements'"),
			mcp.Required(),
		),
		mcp.WithString("file_path",
			mcp.Description("Path to the file to analyze (for 'analyze_file' operation)"),
		),
		mcp.WithString("directory_path",
			mcp.Description("Path to the directory to analyze (for 'analyze_directory' operation)"),
		),
		mcp.WithArray("file_patterns",
			mcp.Description("File patterns to include in the analysis (e.g., ['*.go', '*.js'])"),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("Whether to analyze subdirectories recursively (for 'analyze_directory' operation)"),
		),
		mcp.WithString("target_path",
			mcp.Description("Path to the file or directory to analyze (for 'find_issues' and 'suggest_improvements' operations)"),
		),
		mcp.WithArray("issue_types",
			mcp.Description("Types of issues to look for (e.g., ['complexity', 'duplication', 'naming'])"),
		),
		mcp.WithString("severity",
			mcp.Description("Minimum severity level of issues to report ('low', 'medium', 'high')"),
		),
		mcp.WithArray("improvement_types",
			mcp.Description("Types of improvements to suggest (e.g., ['refactoring', 'performance', 'readability'])"),
		),
	)

	// Wrap the handler with stats tracking
	wrappedHandler := stats.WrapHandler("codeanalysis", HandleCodeAnalysis)

	// Register the tool with the wrapped handler
	mcpServer.AddTool(codeAnalysisTool, wrappedHandler)
}
