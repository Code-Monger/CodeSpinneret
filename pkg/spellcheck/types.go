package spellcheck

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
