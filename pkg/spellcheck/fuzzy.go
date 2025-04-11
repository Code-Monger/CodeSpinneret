package spellcheck

import (
	"bufio"
	"log"
	"strings"

	"github.com/sajari/fuzzy"
)

var fuzzyModel *fuzzy.Model

// initFuzzyModel initializes the fuzzy model with the embedded dictionary
func initFuzzyModel() *fuzzy.Model {
	if fuzzyModel != nil {
		return fuzzyModel
	}

	// Create a new fuzzy model
	model := fuzzy.NewModel()

	// Set the model parameters
	model.SetDepth(2)     // Maximum edit distance
	model.SetThreshold(1) // Minimum frequency threshold
	model.SetUseAutocomplete(true)

	// Open the embedded words.txt file
	file, err := embeddedFS.Open("data/words.txt")
	if err != nil {
		log.Printf("[SpellCheck] Error opening embedded dictionary: %v", err)
		return model
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Read each line and train the model with the word
	wordCount := 0
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			model.TrainWord(strings.ToLower(word))
			wordCount++
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		log.Printf("[SpellCheck] Error reading embedded dictionary: %v", err)
	}

	log.Printf("[SpellCheck] Trained fuzzy model with %d words", wordCount)

	// Train with additional common programming terms
	for word := range getCommonProgrammingTerms() {
		model.TrainWord(word)
	}

	fuzzyModel = model
	return model
}

// getFuzzyModel returns the initialized fuzzy model
func getFuzzyModel() *fuzzy.Model {
	if fuzzyModel == nil {
		fuzzyModel = initFuzzyModel()
	}
	return fuzzyModel
}

// getCommonProgrammingTerms returns a map of common programming terms
func getCommonProgrammingTerms() map[string]bool {
	return map[string]bool{
		"var": true, "func": true, "int": true, "str": true, "bool": true,
		"len": true, "fmt": true, "println": true, "printf": true, "sprintf": true,
		"args": true, "argv": true, "argc": true, "param": true, "params": true,
		"init": true, "exec": true, "eval": true, "impl": true, "pkg": true,
		"lib": true, "src": true, "dest": true, "tmp": true, "temp": true,
		"dir": true, "dirs": true, "cmd": true, "cmds": true, "env": true,
		"config": true, "cfg": true, "ctx": true, "req": true, "res": true,
		"err": true, "stdin": true, "stdout": true, "stderr": true,
		"account": true, "function": true, "variable": true, "string": true,
		"comment": true, "spelling": true, "mistake": true, "package": true,
		"import": true, "code": true, "program": true, "software": true,
		"computer": true, "system": true, "data": true, "file": true,
		"directory": true, "path": true, "name": true, "type": true,
		"value": true, "method": true, "class": true, "object": true,
		"interface": true, "struct": true, "array": true, "slice": true,
		"map": true, "channel": true, "goroutine": true, "thread": true,
		"process": true, "memory": true, "disk": true, "network": true,
		"server": true, "client": true, "request": true, "response": true,
		"error": true, "exception": true, "bug": true, "debug": true,
		"test": true, "unit": true, "integration": true, "performance": true,
		"benchmark": true, "profile": true, "optimize": true, "refactor": true,
		"design": true, "pattern": true, "algorithm": true, "complexity": true,
		"efficiency": true, "speed": true, "cpu_usage": true, "cpu": true,
		"io": true, "input": true, "output": true, "read": true, "write": true,
		"open": true, "close": true, "create": true, "delete": true,
		"update": true, "insert": true, "select": true, "query": true,
		"database": true, "table": true, "row": true, "column": true,
		"field": true, "record": true, "key": true, "val": true,
		"index": true, "search": true, "find": true, "match": true,
		"replace": true, "regex": true, "regular": true, "expression": true,
	}
}
