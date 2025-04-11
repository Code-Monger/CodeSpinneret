package rag

import "time"

// IndexResult represents the result of indexing a repository
type IndexResult struct {
	FilesIndexed    int
	SnippetsIndexed int
	TotalTokens     int
	IndexSize       int64
	TimeTaken       time.Duration
	FileTypes       map[string]int
}

// QueryResult represents the result of querying a repository
type QueryResult struct {
	Results   []CodeSnippet
	TimeTaken time.Duration
}

// CodeSnippet represents a code snippet retrieved from the repository
type CodeSnippet struct {
	FilePath   string
	Snippet    string
	Similarity float64
}

// VectorDB represents a simple vector database for storing and retrieving embeddings
type VectorDB struct {
	Embeddings []Embedding
}

// Embedding represents a vector embedding of a code snippet
type Embedding struct {
	FilePath    string
	Content     string
	Vector      []float64
	Language    string
	LineNumbers string
}
