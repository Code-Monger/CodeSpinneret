package spellcheck

import (
	"bufio"
	"embed"
	"log"
	"strings"
)

//go:embed data/words.txt
var embeddedFS embed.FS

// loadEmbeddedDictionary loads words from the embedded dictionary file
func loadEmbeddedDictionary() map[string]bool {
	// Create a map to store the words
	dictionary := make(map[string]bool)

	// Open the embedded words.txt file
	file, err := embeddedFS.Open("data/words.txt")
	if err != nil {
		log.Printf("[SpellCheck] Error opening embedded dictionary: %v", err)
		return dictionary
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Read each line and add the word to the dictionary
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			dictionary[strings.ToLower(word)] = true
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		log.Printf("[SpellCheck] Error reading embedded dictionary: %v", err)
	}

	return dictionary
}
