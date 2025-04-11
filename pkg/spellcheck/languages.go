package spellcheck

import (
	"strings"
)

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
