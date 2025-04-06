package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rhinocodelab/IntegrityWatchdog/config"
	"github.com/rhinocodelab/IntegrityWatchdog/scanner"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new baseline",
	Long: `Create a new baseline of your file system.
This command will:
1. Create a .fim directory in your home folder
2. Create a configuration file if it doesn't exist
3. Scan configured directories
4. Store file metadata and hashes
5. Create a baseline file at ~/.fim/baseline.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}

		// Create .fim directory
		fimDir := filepath.Join(homeDir, ".fim")
		if err := os.MkdirAll(fimDir, 0755); err != nil {
			return fmt.Errorf("failed to create .fim directory: %v", err)
		}

		// Check if config file exists
		configPath := filepath.Join(fimDir, "fim.conf")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Create default config file
			fmt.Println("Creating default configuration file...")
			if err := createDefaultConfig(configPath); err != nil {
				return fmt.Errorf("failed to create config file: %v", err)
			}
			fmt.Println("Please edit the configuration file at:", configPath)
			fmt.Println("Then run 'fim init' again to create the baseline.")
			return nil
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}

		// Create scanner
		s := scanner.NewScanner(cfg)

		// Scan all configured paths
		fmt.Println("Scanning configured paths...")
		baseline, err := s.ScanPaths()
		if err != nil {
			return fmt.Errorf("failed to scan paths: %v", err)
		}

		// Save baseline
		baselinePath := filepath.Join(fimDir, "baseline.json")
		if err := baseline.Save(baselinePath); err != nil {
			return fmt.Errorf("failed to save baseline: %v", err)
		}

		fmt.Printf("Baseline created successfully at %s\n", baselinePath)
		return nil
	},
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configPath string) error {
	// Default config content
	configContent := `[monitor]
# Required: List of paths to monitor (comma-separated)
paths = /etc, /usr/bin

# Optional: Paths to exclude from monitoring (comma-separated)
exclude = /tmp, /var/log, /proc

[logging]
# Optional: Log file path
logfile = /var/log/fim.log

[output]
# Optional: Enable verbose output
verbose = true
`

	// Write config file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
