package rag

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

// createVectorDB creates a new vector database
func createVectorDB() *VectorDB {
	return &VectorDB{
		Embeddings: []Embedding{},
	}
}

// addEmbedding adds an embedding to the vector database
func (db *VectorDB) addEmbedding(filePath, content, language, lineNumbers string, vector []float64) {
	db.Embeddings = append(db.Embeddings, Embedding{
		FilePath:    filePath,
		Content:     content,
		Vector:      vector,
		Language:    language,
		LineNumbers: lineNumbers,
	})
}

// search searches the vector database for similar embeddings
func (db *VectorDB) search(queryVector []float64, numResults int) []CodeSnippet {
	if len(db.Embeddings) == 0 {
		return []CodeSnippet{}
	}

	// Calculate similarity scores
	type ScoredEmbedding struct {
		Embedding  Embedding
		Similarity float64
	}
	var scoredEmbeddings []ScoredEmbedding

	for _, embedding := range db.Embeddings {
		similarity := cosineSimilarity(queryVector, embedding.Vector)
		scoredEmbeddings = append(scoredEmbeddings, ScoredEmbedding{
			Embedding:  embedding,
			Similarity: similarity,
		})
	}

	// Sort by similarity (descending)
	// In a real implementation, we would sort the embeddings by similarity
	// For demonstration purposes, we'll just return random embeddings

	// Convert to CodeSnippets
	var results []CodeSnippet
	for i := 0; i < min(numResults, len(scoredEmbeddings)); i++ {
		results = append(results, CodeSnippet{
			FilePath:   scoredEmbeddings[i].Embedding.FilePath,
			Snippet:    scoredEmbeddings[i].Embedding.Content,
			Similarity: scoredEmbeddings[i].Similarity,
		})
	}

	return results
}

// saveToFile saves the vector database to a file
func (db *VectorDB) saveToFile(filePath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	// Marshal to JSON
	data, err := json.Marshal(db)
	if err != nil {
		return fmt.Errorf("error marshaling vector database: %v", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("error writing vector database: %v", err)
	}

	return nil
}

// loadFromFile loads the vector database from a file
func loadVectorDBFromFile(filePath string) (*VectorDB, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading vector database: %v", err)
	}

	// Unmarshal from JSON
	var db VectorDB
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("error unmarshaling vector database: %v", err)
	}

	return &db, nil
}

// generateEmbedding generates an embedding for a text
func generateEmbedding(text string) []float64 {
	// In a real implementation, we would use an embedding model
	// For demonstration purposes, we'll just return a random vector
	vector := make([]float64, 128)
	for i := range vector {
		vector[i] = rand.Float64()
	}
	return vector
}

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, magnitudeA, magnitudeB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		magnitudeA += a[i] * a[i]
		magnitudeB += b[i] * b[i]
	}

	magnitudeA = math.Sqrt(magnitudeA)
	magnitudeB = math.Sqrt(magnitudeB)

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0
	}

	return dotProduct / (magnitudeA * magnitudeB)
}
