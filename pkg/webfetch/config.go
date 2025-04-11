package webfetch

import (
	"sync"
)

// Config represents the configuration for the webfetch service
type Config struct {
	// UserAgent is the User-Agent header to use for requests
	UserAgent string

	// DefaultTimeout is the default timeout in seconds
	DefaultTimeout int

	// MaxContentSize is the maximum size of content to fetch in bytes
	MaxContentSize int
}

var (
	// defaultConfig is the default configuration
	defaultConfig = Config{
		UserAgent:      "CodeSpinneret-WebFetch/1.0",
		DefaultTimeout: 30,
		MaxContentSize: 1024 * 1024, // 1MB
	}

	// configMutex protects the config
	configMutex sync.RWMutex

	// config is the current configuration
	config = defaultConfig
)

// GetConfig returns the current configuration
func GetConfig() Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return config
}

// SetConfig sets the configuration
func SetConfig(newConfig Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	config = newConfig
}

// ResetConfig resets the configuration to the default
func ResetConfig() {
	configMutex.Lock()
	defer configMutex.Unlock()
	config = defaultConfig
}
