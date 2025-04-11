package spellcheck

import (
	"strings"
	"unicode"
)

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

// isVowel checks if a character is a vowel
func isVowel(c rune) bool {
	return strings.ContainsRune("aeiouAEIOU", c)
}
