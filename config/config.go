package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the FIM tool configuration
type Config struct {
	Monitor struct {
		Paths   []string `mapstructure:"paths"`
		Exclude []string `mapstructure:"exclude"`
	} `mapstructure:"monitor"`
	Logging struct {
		LogFile string `mapstructure:"logfile"`
	} `mapstructure:"logging"`
	Output struct {
		Verbose bool `mapstructure:"verbose"`
	} `mapstructure:"output"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}

	// Set default exclude paths
	cfg.Monitor.Exclude = []string{
		"/tmp",
		"/var/log",
	}

	// Set default log file
	cfg.Logging.LogFile = "/var/log/fim.log"

	// Set default output settings
	cfg.Output.Verbose = true

	return cfg
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	// Return the path to the config file in the .fim directory
	return filepath.Join(homeDir, ".fim", "fim.conf"), nil
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	// Check if monitor paths are specified
	if len(c.Monitor.Paths) == 0 {
		return fmt.Errorf("no monitor paths specified in fim.conf. Please add [monitor] section with paths")
	}

	// Validate and clean paths
	for i, path := range c.Monitor.Paths {
		// Remove trailing slashes
		c.Monitor.Paths[i] = strings.TrimRight(path, "/")

		// Check if path exists
		if _, err := os.Stat(c.Monitor.Paths[i]); os.IsNotExist(err) {
			return fmt.Errorf("monitor path does not exist: %s", path)
		}
	}

	// Validate log file path if specified
	if c.Logging.LogFile != "" {
		// Check if the directory exists
		logDir := filepath.Dir(c.Logging.LogFile)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			return fmt.Errorf("log directory does not exist: %s", logDir)
		}
	}

	return nil
}

// LoadConfig loads the configuration from fim.conf
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Get config file path
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found at %s", configPath)
	}

	// Set up Viper
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("ini")

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Unmarshal the config
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	// Validate the config
	if err := cfg.ValidateConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// IsExcluded checks if a path should be excluded from monitoring
func (c *Config) IsExcluded(path string) bool {
	for _, exclude := range c.Monitor.Exclude {
		if matched, _ := filepath.Match(exclude, path); matched {
			return true
		}
	}
	return false
}
