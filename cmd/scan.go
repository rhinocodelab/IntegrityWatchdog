package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rhinocodelab/IntegrityWatchdog/config"
	"github.com/rhinocodelab/IntegrityWatchdog/daemon"
	"github.com/rhinocodelab/IntegrityWatchdog/scanner"
	"github.com/rhinocodelab/IntegrityWatchdog/storage"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	interval   string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for changes",
	Long: `Scan the configured paths for changes since the last baseline.
This command will compare the current state of files with the baseline
and report any changes detected.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if configuration file exists
		configPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println("Configuration file not found at:", configPath)
			fmt.Println("Please run 'fim init' to create a configuration file.")
			return fmt.Errorf("configuration file not found")
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %v", err)
		}

		// Check if daemon mode is requested
		if daemonMode {
			// Check if daemon is already running
			if daemon.IsRunning() {
				return fmt.Errorf("FIM daemon is already running")
			}

			// Parse interval
			var scanInterval time.Duration
			if interval != "" {
				var err error
				scanInterval, err = time.ParseDuration(interval)
				if err != nil {
					return fmt.Errorf("invalid interval format: %v", err)
				}
			} else {
				// Default interval: 5 minutes
				scanInterval = 5 * time.Minute
			}

			// Create and start daemon
			d, err := daemon.NewDaemon(cfg, scanInterval)
			if err != nil {
				return fmt.Errorf("failed to create daemon: %v", err)
			}

			if err := d.Start(); err != nil {
				return fmt.Errorf("failed to start daemon: %v", err)
			}

			fmt.Printf("FIM daemon started with scan interval: %s\n", scanInterval)
			return nil
		}

		// Regular scan mode
		// Load baseline
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		baselinePath := filepath.Join(homeDir, ".fim", "baseline.json")

		baseline, err := storage.Load(baselinePath)
		if err != nil {
			return fmt.Errorf("failed to load baseline: %v", err)
		}

		// Create scanner
		s := scanner.NewScanner(cfg)

		// Print what we're scanning
		fmt.Println("Scanning configured paths...")
		for _, path := range cfg.Monitor.Paths {
			fmt.Printf("  - %s\n", path)
			// Check if it's a symlink
			if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
				realPath, err := filepath.EvalSymlinks(path)
				if err == nil {
					fmt.Printf("    (symlink to: %s)\n", realPath)
				}
			}
		}

		// Scan paths
		currentState, err := s.ScanPaths()
		if err != nil {
			return fmt.Errorf("failed to scan paths: %v", err)
		}

		// Compare with baseline
		changes := baseline.Compare(currentState)

		// Output results
		if jsonOutput {
			// JSON output
			jsonData, err := json.MarshalIndent(changes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal changes to JSON: %v", err)
			}
			fmt.Println(string(jsonData))
		} else {
			// Text output
			if len(changes.Added) == 0 && len(changes.Modified) == 0 && len(changes.Deleted) == 0 {
				fmt.Println("No changes detected.")
			} else {
				fmt.Println("\nChanges detected:")
				for _, file := range changes.Added {
					fmt.Printf("[+] %s\n", file.Path)
				}
				for _, file := range changes.Modified {
					fmt.Printf("[*] %s\n", file.Path)
				}
				for _, file := range changes.Deleted {
					fmt.Printf("[-] %s\n", file.Path)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	scanCmd.Flags().StringVar(&interval, "interval", "", "Scan interval in daemon mode (e.g., 5m, 1h)")
}
