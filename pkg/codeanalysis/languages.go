package codeanalysis

import (
	"go/ast"
	"regexp"
	"strings"
)

// analyzeJSFile analyzes a JavaScript/TypeScript file
func analyzeJSFile(content string, result *FileAnalysisResult) {
	// Count functions
	functionRegex := regexp.MustCompile(`(function\s+\w+\s*\(|const\s+\w+\s*=\s*function\s*\(|const\s+\w+\s*=\s*\(.*\)\s*=>|class\s+\w+)`)
	functionMatches := functionRegex.FindAllString(content, -1)
	result.FunctionCount = len(functionMatches)

	// Count classes
	classRegex := regexp.MustCompile(`class\s+\w+`)
	classMatches := classRegex.FindAllString(content, -1)
	result.ClassCount = len(classMatches)

	// Count comments
	commentRegex := regexp.MustCompile(`(//.*|/\*[\s\S]*?\*/)`)
	commentMatches := commentRegex.FindAllString(content, -1)
	result.CommentLines = len(commentMatches)

	// Extract imports/dependencies
	importRegex := regexp.MustCompile(`(import\s+.*?from\s+['"](.+?)['"]|require\s*\(\s*['"](.+?)['"]\s*\))`)
	importMatches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range importMatches {
		if match[2] != "" {
			result.Dependencies = append(result.Dependencies, match[2])
		} else if match[3] != "" {
			result.Dependencies = append(result.Dependencies, match[3])
		}
	}

	// Find top complex functions (simplified)
	functionComplexityRegex := regexp.MustCompile(`(function\s+(\w+)|const\s+(\w+)\s*=\s*function|const\s+(\w+)\s*=\s*\(.*\)\s*=>)`)
	functionComplexityMatches := functionComplexityRegex.FindAllStringSubmatch(content, -1)

	for i, match := range functionComplexityMatches {
		if i >= 5 {
			break
		}

		var functionName string
		if match[2] != "" {
			functionName = match[2]
		} else if match[3] != "" {
			functionName = match[3]
		} else if match[4] != "" {
			functionName = match[4]
		} else {
			functionName = "anonymous_" + string(i)
		}

		// Find line number (simplified)
		lines := strings.Split(content, "\n")
		lineNumber := 0
		for j, line := range lines {
			if strings.Contains(line, match[0]) {
				lineNumber = j + 1
				break
			}
		}

		// Calculate complexity (simplified)
		complexity := 1.0

		// Count if statements
		ifCount := strings.Count(match[0], "if ")
		complexity += float64(ifCount) * 0.5

		// Count loops
		loopCount := strings.Count(match[0], "for ") + strings.Count(match[0], "while ")
		complexity += float64(loopCount) * 0.7

		// Count switch statements
		switchCount := strings.Count(match[0], "switch ")
		complexity += float64(switchCount) * 0.6

		result.TopFunctions = append(result.TopFunctions, FunctionInfo{
			Name:       functionName,
			Line:       lineNumber,
			Complexity: complexity,
		})
	}

	// Calculate overall complexity (simplified)
	result.ComplexityScore = 1.0
	if len(result.TopFunctions) > 0 {
		totalComplexity := 0.0
		for _, fn := range result.TopFunctions {
			totalComplexity += fn.Complexity
		}
		result.ComplexityScore = totalComplexity / float64(len(result.TopFunctions))
	}
}

// analyzePythonFile analyzes a Python file
func analyzePythonFile(content string, result *FileAnalysisResult) {
	// Count functions
	functionRegex := regexp.MustCompile(`def\s+\w+\s*\(`)
	functionMatches := functionRegex.FindAllString(content, -1)
	result.FunctionCount = len(functionMatches)

	// Count classes
	classRegex := regexp.MustCompile(`class\s+\w+`)
	classMatches := classRegex.FindAllString(content, -1)
	result.ClassCount = len(classMatches)

	// Count comments
	commentRegex := regexp.MustCompile(`(#.*|'''[\s\S]*?'''|"""[\s\S]*?""")`)
	commentMatches := commentRegex.FindAllString(content, -1)
	result.CommentLines = len(commentMatches)

	// Extract imports/dependencies
	importRegex := regexp.MustCompile(`(import\s+(\w+)|from\s+(\w+)(?:\.\w+)*\s+import)`)
	importMatches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range importMatches {
		if match[2] != "" {
			result.Dependencies = append(result.Dependencies, match[2])
		} else if match[3] != "" {
			result.Dependencies = append(result.Dependencies, match[3])
		}
	}

	// Calculate complexity (simplified)
	result.ComplexityScore = 1.0

	// Count if statements
	ifCount := strings.Count(content, "if ")
	result.ComplexityScore += float64(ifCount) * 0.1

	// Count loops
	loopCount := strings.Count(content, "for ") + strings.Count(content, "while ")
	result.ComplexityScore += float64(loopCount) * 0.2

	// Count exception handling
	exceptionCount := strings.Count(content, "try:") + strings.Count(content, "except ")
	result.ComplexityScore += float64(exceptionCount) * 0.3
}

// analyzeJavaFile analyzes a Java file
func analyzeJavaFile(content string, result *FileAnalysisResult) {
	// Count methods
	methodRegex := regexp.MustCompile(`(public|private|protected|static|\s) +[\w\<\>\[\]]+\s+(\w+) *\([^\)]*\) *(\{?|[^;])`)
	methodMatches := methodRegex.FindAllString(content, -1)
	result.FunctionCount = len(methodMatches)

	// Count classes
	classRegex := regexp.MustCompile(`(class|interface|enum)\s+\w+`)
	classMatches := classRegex.FindAllString(content, -1)
	result.ClassCount = len(classMatches)

	// Count comments
	commentRegex := regexp.MustCompile(`(//.*|/\*[\s\S]*?\*/)`)
	commentMatches := commentRegex.FindAllString(content, -1)
	result.CommentLines = len(commentMatches)

	// Extract imports
	importRegex := regexp.MustCompile(`import\s+([^;]+);`)
	importMatches := importRegex.FindAllStringSubmatch(content, -1)
	for _, match := range importMatches {
		if match[1] != "" {
			result.Dependencies = append(result.Dependencies, match[1])
		}
	}

	// Calculate complexity (simplified)
	result.ComplexityScore = 1.0

	// Count if statements
	ifCount := strings.Count(content, "if ")
	result.ComplexityScore += float64(ifCount) * 0.1

	// Count loops
	loopCount := strings.Count(content, "for ") + strings.Count(content, "while ")
	result.ComplexityScore += float64(loopCount) * 0.2

	// Count switch statements
	switchCount := strings.Count(content, "switch ")
	result.ComplexityScore += float64(switchCount) * 0.3

	// Count exception handling
	exceptionCount := strings.Count(content, "try ") + strings.Count(content, "catch ")
	result.ComplexityScore += float64(exceptionCount) * 0.4
}

// analyzeCFile analyzes a C/C++ file
func analyzeCFile(content string, result *FileAnalysisResult) {
	// Count functions
	functionRegex := regexp.MustCompile(`\w+\s+\w+\s*\([^{]*\)\s*\{`)
	functionMatches := functionRegex.FindAllString(content, -1)
	result.FunctionCount = len(functionMatches)

	// Count structs/classes
	structRegex := regexp.MustCompile(`(struct|class)\s+\w+`)
	structMatches := structRegex.FindAllString(content, -1)
	result.ClassCount = len(structMatches)

	// Count comments
	commentRegex := regexp.MustCompile(`(//.*|/\*[\s\S]*?\*/)`)
	commentMatches := commentRegex.FindAllString(content, -1)
	result.CommentLines = len(commentMatches)

	// Extract includes
	includeRegex := regexp.MustCompile(`#include\s+[<"]([^>"]+)[>"]`)
	includeMatches := includeRegex.FindAllStringSubmatch(content, -1)
	for _, match := range includeMatches {
		if match[1] != "" {
			result.Dependencies = append(result.Dependencies, match[1])
		}
	}

	// Calculate complexity (simplified)
	result.ComplexityScore = 1.0

	// Count if statements
	ifCount := strings.Count(content, "if ")
	result.ComplexityScore += float64(ifCount) * 0.1

	// Count loops
	loopCount := strings.Count(content, "for ") + strings.Count(content, "while ")
	result.ComplexityScore += float64(loopCount) * 0.2

	// Count switch statements
	switchCount := strings.Count(content, "switch ")
	result.ComplexityScore += float64(switchCount) * 0.3

	// Count goto statements
	gotoCount := strings.Count(content, "goto ")
	result.ComplexityScore += float64(gotoCount) * 0.5
}

// analyzeCSharpFile analyzes a C# file
func analyzeCSharpFile(content string, result *FileAnalysisResult) {
	// Count methods
	methodRegex := regexp.MustCompile(`(public|private|protected|internal|static|\s) +[\w\<\>\[\]]+\s+(\w+) *\([^\)]*\) *(\{?|[^;])`)
	methodMatches := methodRegex.FindAllString(content, -1)
	result.FunctionCount = len(methodMatches)

	// Count classes
	classRegex := regexp.MustCompile(`(class|interface|enum|struct)\s+\w+`)
	classMatches := classRegex.FindAllString(content, -1)
	result.ClassCount = len(classMatches)

	// Count comments
	commentRegex := regexp.MustCompile(`(//.*|/\*[\s\S]*?\*/)`)
	commentMatches := commentRegex.FindAllString(content, -1)
	result.CommentLines = len(commentMatches)

	// Extract using statements
	usingRegex := regexp.MustCompile(`using\s+([^;]+);`)
	usingMatches := usingRegex.FindAllStringSubmatch(content, -1)
	for _, match := range usingMatches {
		if match[1] != "" {
			result.Dependencies = append(result.Dependencies, match[1])
		}
	}

	// Calculate complexity (simplified)
	result.ComplexityScore = 1.0

	// Count if statements
	ifCount := strings.Count(content, "if ")
	result.ComplexityScore += float64(ifCount) * 0.1

	// Count loops
	loopCount := strings.Count(content, "for ") + strings.Count(content, "while ") + strings.Count(content, "foreach ")
	result.ComplexityScore += float64(loopCount) * 0.2

	// Count switch statements
	switchCount := strings.Count(content, "switch ")
	result.ComplexityScore += float64(switchCount) * 0.3

	// Count exception handling
	exceptionCount := strings.Count(content, "try ") + strings.Count(content, "catch ")
	result.ComplexityScore += float64(exceptionCount) * 0.4
}

// analyzeGenericFile analyzes a file with unknown language
func analyzeGenericFile(content string, result *FileAnalysisResult) {
	// Count lines of code (already done in analyzeFile)

	// Count comments (simplified)
	commentRegex := regexp.MustCompile(`(//.*|/\*[\s\S]*?\*/|#.*)`)
	commentMatches := commentRegex.FindAllString(content, -1)
	result.CommentLines = len(commentMatches)

	// Calculate complexity (very simplified)
	result.ComplexityScore = 1.0

	// Count if statements
	ifCount := strings.Count(content, "if ")
	result.ComplexityScore += float64(ifCount) * 0.1

	// Count loops
	loopCount := strings.Count(content, "for ") + strings.Count(content, "while ")
	result.ComplexityScore += float64(loopCount) * 0.2
}

// calculateGoFunctionComplexity calculates the complexity of a Go function
func calculateGoFunctionComplexity(funcDecl *ast.FuncDecl) float64 {
	complexity := 1.0

	// Count statements that increase complexity
	ast.Inspect(funcDecl, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.IfStmt:
			complexity += 0.5
		case *ast.ForStmt, *ast.RangeStmt:
			complexity += 0.7
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			complexity += 0.6
		case *ast.SelectStmt:
			complexity += 0.8
		case *ast.GoStmt:
			complexity += 0.5
		case *ast.DeferStmt:
			complexity += 0.3
		}
		return true
	})

	return complexity
}
