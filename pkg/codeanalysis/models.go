package codeanalysis

import "time"

// FileAnalysisResult represents the result of analyzing a single file
type FileAnalysisResult struct {
	Language        string
	LinesOfCode     int
	FunctionCount   int
	ClassCount      int
	CommentLines    int
	ComplexityScore float64
	TopFunctions    []FunctionInfo
	Dependencies    []string
	TimeTaken       time.Duration
}

// FunctionInfo represents information about a function
type FunctionInfo struct {
	Name       string
	Line       int
	Complexity float64
}

// DirectoryAnalysisResult represents the result of analyzing a directory
type DirectoryAnalysisResult struct {
	FilesAnalyzed      int
	TotalLinesOfCode   int
	TotalFunctionCount int
	TotalClassCount    int
	TotalCommentLines  int
	AverageComplexity  float64
	LanguageBreakdown  map[string]int
	TopComplexFiles    []FileInfo
	TimeTaken          time.Duration
}

// FileInfo represents information about a file
type FileInfo struct {
	Path        string
	LinesOfCode int
	Complexity  float64
}

// IssuesResult represents the result of finding issues
type IssuesResult struct {
	TotalIssues  int
	IssuesByType map[string]int
	Issues       []IssueInfo
	TimeTaken    time.Duration
}

// IssueInfo represents information about an issue
type IssueInfo struct {
	Type     string
	Message  string
	FilePath string
	Line     int
	Snippet  string
}

// ImprovementsResult represents the result of suggesting improvements
type ImprovementsResult struct {
	TotalSuggestions  int
	SuggestionsByType map[string]int
	Suggestions       []SuggestionInfo
	TimeTaken         time.Duration
}

// SuggestionInfo represents information about a suggestion
type SuggestionInfo struct {
	Type        string
	Title       string
	Description string
	FilePath    string
	Line        int
	Before      string
	After       string
}
