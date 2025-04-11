package spellcheck

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

// isMisspelled checks if a word is misspelled
func isMisspelled(word, dictionaryType string) bool {
	// This is an enhanced implementation with better spell checking using the fuzzy model

	// Convert to lowercase for case-insensitive comparison
	lowerWord := strings.ToLower(word)

	// Skip short words (likely not misspelled)
	if len(lowerWord) <= 2 {
		return false
	}

	// Check against common programming terms
	if isCommonProgrammingTerm(lowerWord) {
		return false
	}

	// Check against common English words
	commonWords := getCommonWords()
	if commonWords[lowerWord] {
		return false
	}

	// Check for common misspellings
	commonMisspellings := getCommonMisspellings()
	if _, ok := commonMisspellings[lowerWord]; ok {
		return true
	}

	// Check for common programming misspellings
	programmingMisspellings := getProgrammingMisspellings()
	if _, ok := programmingMisspellings[lowerWord]; ok {
		return true
	}

	// Additional heuristics for detecting misspellings

	// Check for repeated letters (more than 2 of the same letter in a row)
	for i := 0; i < len(lowerWord)-2; i++ {
		if lowerWord[i] == lowerWord[i+1] && lowerWord[i] == lowerWord[i+2] {
			return true
		}
	}

	// Check for unlikely letter combinations
	unlikelyCombinations := []string{"tch", "mpt", "thm", "rld", "lld", "gth", "kth"}
	for _, combo := range unlikelyCombinations {
		if strings.Contains(lowerWord, combo) {
			// These are actually valid in English, so we'll check against our dictionary
			if !commonWords[lowerWord] {
				return true
			}
		}
	}

	// If the word is not in our dictionary and doesn't match any of our rules,
	// we'll consider it misspelled
	return !commonWords[lowerWord]
}

// getSuggestions gets spelling suggestions for a misspelled word
func getSuggestions(word, dictionaryType string) []string {
	// Use the fuzzy model to get suggestions
	model := getFuzzyModel()

	// Get suggestions from the fuzzy model
	suggestions := model.SpellCheckSuggestions(strings.ToLower(word), 5)

	// If we didn't get any suggestions from the fuzzy model, fall back to our hardcoded suggestions
	if len(suggestions) == 0 {
		// Check for common misspellings
		commonMisspellings := map[string][]string{
			"coment":    {"comment"},
			"speling":   {"spelling"},
			"mesage":    {"message"},
			"acount":    {"account"},
			"messge":    {"message"},
			"mispelled": {"misspelled"},
		}

		lowerWord := strings.ToLower(word)
		if hardcodedSuggestions, ok := commonMisspellings[lowerWord]; ok {
			return hardcodedSuggestions
		}

		// Otherwise, generate some simple suggestions
		var fallbackSuggestions []string

		// Add 'e' if the word ends with a consonant
		if len(word) > 2 && !isVowel(rune(word[len(word)-1])) {
			fallbackSuggestions = append(fallbackSuggestions, word+"e")
		}

		// Double the last letter
		if len(word) > 2 {
			fallbackSuggestions = append(fallbackSuggestions, word+string(word[len(word)-1]))
		}

		// Add common suffixes
		fallbackSuggestions = append(fallbackSuggestions, word+"s", word+"ed", word+"ing")

		return fallbackSuggestions
	}

	return suggestions
}
