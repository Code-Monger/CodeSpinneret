package codeanalysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// analyzeFile analyzes a single file
func analyzeFile(filePath string) (*FileAnalysisResult, error) {
	startTime := time.Now()

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		return nil, err
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Determine language based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := getLanguageFromExtension(ext)

	// Count lines of code
	lines := strings.Split(string(content), "\n")
	linesOfCode := len(lines)

	// Initialize result
	result := &FileAnalysisResult{
		Language:    language,
		LinesOfCode: linesOfCode,
		TimeTaken:   time.Since(startTime),
	}

	// Analyze based on language
	switch language {
	case "Go":
		analyzeGoFile(filePath, string(content), result)
	case "JavaScript", "TypeScript":
		analyzeJSFile(string(content), result)
	case "Python":
		analyzePythonFile(string(content), result)
	case "Java":
		analyzeJavaFile(string(content), result)
	case "C", "C++":
		analyzeCFile(string(content), result)
	case "C#":
		analyzeCSharpFile(string(content), result)
	default:
		// Generic analysis for other languages
		analyzeGenericFile(string(content), result)
	}

	return result, nil
}

// analyzeDirectory analyzes a directory of code files
func analyzeDirectory(dirPath string, filePatterns []string, recursive bool) (*DirectoryAnalysisResult, error) {
	startTime := time.Now()

	// Check if directory exists
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}
	if !dirInfo.IsDir() {
		return nil, err
	}

	// Initialize result
	result := &DirectoryAnalysisResult{
		LanguageBreakdown: make(map[string]int),
		TopComplexFiles:   []FileInfo{},
	}

	// Find all files matching the patterns
	var files []string
	for _, pattern := range filePatterns {
		walkFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip directories
			if info.IsDir() {
				if !recursive && path != dirPath {
					return filepath.SkipDir
				}
				return nil
			}

			// Check if file matches the pattern
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				files = append(files, path)
			}

			return nil
		}

		err := filepath.Walk(dirPath, walkFunc)
		if err != nil {
			return nil, err
		}
	}

	// Analyze each file
	var wg sync.WaitGroup
	var mu sync.Mutex
	fileResults := make(map[string]*FileAnalysisResult)

	for _, file := range files {
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			fileResult, err := analyzeFile(filePath)
			if err != nil {
				return
			}

			mu.Lock()
			fileResults[filePath] = fileResult
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	// Aggregate results
	result.FilesAnalyzed = len(fileResults)
	for filePath, fileResult := range fileResults {
		result.TotalLinesOfCode += fileResult.LinesOfCode
		result.TotalFunctionCount += fileResult.FunctionCount
		result.TotalClassCount += fileResult.ClassCount
		result.TotalCommentLines += fileResult.CommentLines

		result.LanguageBreakdown[fileResult.Language]++

		// Add to top complex files
		result.TopComplexFiles = append(result.TopComplexFiles, FileInfo{
			Path:        filePath,
			LinesOfCode: fileResult.LinesOfCode,
			Complexity:  fileResult.ComplexityScore,
		})
	}

	// Calculate average complexity
	if result.FilesAnalyzed > 0 {
		totalComplexity := 0.0
		for _, fileResult := range fileResults {
			totalComplexity += fileResult.ComplexityScore
		}
		result.AverageComplexity = totalComplexity / float64(result.FilesAnalyzed)
	}

	// Sort top complex files by complexity (descending)
	sort.Slice(result.TopComplexFiles, func(i, j int) bool {
		return result.TopComplexFiles[i].Complexity > result.TopComplexFiles[j].Complexity
	})

	// Limit to top 10
	if len(result.TopComplexFiles) > 10 {
		result.TopComplexFiles = result.TopComplexFiles[:10]
	}

	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// findIssues finds issues in code
func findIssues(targetPath string, issueTypes []string, severityLevel string) (*IssuesResult, error) {
	startTime := time.Now()

	// Initialize result
	result := &IssuesResult{
		IssuesByType: make(map[string]int),
		Issues:       []IssueInfo{},
	}

	// Check if target is a file or directory
	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		// Analyze directory
		dirResult, err := analyzeDirectory(targetPath, []string{"*.go", "*.js", "*.ts", "*.py", "*.java", "*.c", "*.cpp", "*.h", "*.cs"}, true)
		if err != nil {
			return nil, err
		}

		// Find issues in each file
		for _, fileInfo := range dirResult.TopComplexFiles {
			fileIssues := findFileIssues(fileInfo.Path, issueTypes, severityLevel)
			for _, issue := range fileIssues {
				result.Issues = append(result.Issues, issue)
				result.IssuesByType[issue.Type]++
			}
		}
	} else {
		// Analyze single file
		fileIssues := findFileIssues(targetPath, issueTypes, severityLevel)
		for _, issue := range fileIssues {
			result.Issues = append(result.Issues, issue)
			result.IssuesByType[issue.Type]++
		}
	}

	result.TotalIssues = len(result.Issues)
	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// findFileIssues finds issues in a single file
func findFileIssues(filePath string, issueTypes []string, severityLevel string) []IssueInfo {
	var issues []IssueInfo

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return issues
	}

	// Determine language based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := getLanguageFromExtension(ext)

	// Find issues based on language and issue types
	for _, issueType := range issueTypes {
		switch issueType {
		case "complexity":
			complexityIssues := findComplexityIssues(filePath, string(content), language, severityLevel)
			issues = append(issues, complexityIssues...)
		case "duplication":
			duplicationIssues := findDuplicationIssues(filePath, string(content), language, severityLevel)
			issues = append(issues, duplicationIssues...)
		case "naming":
			namingIssues := findNamingIssues(filePath, string(content), language, severityLevel)
			issues = append(issues, namingIssues...)
		case "comments":
			commentIssues := findCommentIssues(filePath, string(content), language, severityLevel)
			issues = append(issues, commentIssues...)
		case "unused":
			unusedIssues := findUnusedIssues(filePath, string(content), language, severityLevel)
			issues = append(issues, unusedIssues...)
		}
	}

	return issues
}

// suggestImprovements suggests improvements for code
func suggestImprovements(targetPath string, improvementTypes []string) (*ImprovementsResult, error) {
	startTime := time.Now()

	// Initialize result
	result := &ImprovementsResult{
		SuggestionsByType: make(map[string]int),
		Suggestions:       []SuggestionInfo{},
	}

	// Check if target is a file or directory
	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		// Analyze directory
		dirResult, err := analyzeDirectory(targetPath, []string{"*.go", "*.js", "*.ts", "*.py", "*.java", "*.c", "*.cpp", "*.h", "*.cs"}, true)
		if err != nil {
			return nil, err
		}

		// Suggest improvements for each file
		for _, fileInfo := range dirResult.TopComplexFiles {
			fileSuggestions := suggestFileImprovements(fileInfo.Path, improvementTypes)
			for _, suggestion := range fileSuggestions {
				result.Suggestions = append(result.Suggestions, suggestion)
				result.SuggestionsByType[suggestion.Type]++
			}
		}
	} else {
		// Analyze single file
		fileSuggestions := suggestFileImprovements(targetPath, improvementTypes)
		for _, suggestion := range fileSuggestions {
			result.Suggestions = append(result.Suggestions, suggestion)
			result.SuggestionsByType[suggestion.Type]++
		}
	}

	result.TotalSuggestions = len(result.Suggestions)
	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// suggestFileImprovements suggests improvements for a single file
func suggestFileImprovements(filePath string, improvementTypes []string) []SuggestionInfo {
	var suggestions []SuggestionInfo

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return suggestions
	}

	// Determine language based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	language := getLanguageFromExtension(ext)

	// Suggest improvements based on language and improvement types
	for _, improvementType := range improvementTypes {
		switch improvementType {
		case "refactoring":
			refactoringSuggestions := suggestRefactoring(filePath, string(content), language)
			suggestions = append(suggestions, refactoringSuggestions...)
		case "performance":
			performanceSuggestions := suggestPerformanceImprovements(filePath, string(content), language)
			suggestions = append(suggestions, performanceSuggestions...)
		case "readability":
			readabilitySuggestions := suggestReadabilityImprovements(filePath, string(content), language)
			suggestions = append(suggestions, readabilitySuggestions...)
		case "maintainability":
			maintainabilitySuggestions := suggestMaintainabilityImprovements(filePath, string(content), language)
			suggestions = append(suggestions, maintainabilitySuggestions...)
		}
	}

	return suggestions
}

// analyzeGoFile analyzes a Go file
func analyzeGoFile(filePath, content string, result *FileAnalysisResult) {
	// Parse the Go file
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		// If parsing fails, fall back to generic analysis
		analyzeGenericFile(content, result)
		return
	}

	// Count functions and methods
	functionCount := 0
	typeCount := 0
	commentLines := 0
	topFunctions := []FunctionInfo{}

	// Count comments
	for _, commentGroup := range f.Comments {
		commentLines += len(commentGroup.List)
	}

	// Visit all nodes in the AST
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			functionCount++

			// Calculate function complexity (simplified)
			complexity := calculateGoFunctionComplexity(x)

			// Add to top functions
			topFunctions = append(topFunctions, FunctionInfo{
				Name:       x.Name.Name,
				Line:       fset.Position(x.Pos()).Line,
				Complexity: complexity,
			})
		case *ast.TypeSpec:
			if _, isStruct := x.Type.(*ast.StructType); isStruct {
				typeCount++
			}
		}
		return true
	})

	// Extract imports
	var dependencies []string
	for _, imp := range f.Imports {
		if imp.Path != nil {
			path := strings.Trim(imp.Path.Value, "\"")
			dependencies = append(dependencies, path)
		}
	}

	// Sort top functions by complexity (descending)
	sort.Slice(topFunctions, func(i, j int) bool {
		return topFunctions[i].Complexity > topFunctions[j].Complexity
	})

	// Limit to top 5
	if len(topFunctions) > 5 {
		topFunctions = topFunctions[:5]
	}

	// Calculate overall complexity (simplified)
	complexityScore := float64(functionCount) * 0.1
	for _, fn := range topFunctions {
		complexityScore += fn.Complexity
	}
	complexityScore = complexityScore / float64(max(1, functionCount))

	// Update result
	result.FunctionCount = functionCount
	result.ClassCount = typeCount
	result.CommentLines = commentLines
	result.ComplexityScore = complexityScore
	result.TopFunctions = topFunctions
	result.Dependencies = dependencies
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
