package websearch

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// Config holds the configuration for the websearch service
type Config struct {
	// Search engine URLs
	DuckDuckGoURL string
	BingURL       string
	GoogleURL     string

	// User agent to use for requests
	UserAgent string
}

var (
	config     *Config
	configOnce sync.Once
)

// GetConfig returns the configuration for the websearch service
func GetConfig() *Config {
	configOnce.Do(func() {
		// Load .env file if it exists
		loadEnvFile(".env")

		config = &Config{
			// Search engine URLs
			DuckDuckGoURL: getEnv("DUCKDUCKGO_URL", "https://duckduckgo.com/html"),
			BingURL:       getEnv("BING_URL", "https://www.bing.com/search"),
			GoogleURL:     getEnv("GOOGLE_URL", "https://www.google.com/search"),

			// User agent
			UserAgent: getEnv("USER_AGENT", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		}
	})

	return config
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		// Split by first equals sign
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set environment variable if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
