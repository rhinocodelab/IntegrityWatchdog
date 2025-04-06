package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prashantpokhriyal/fim-tool/daemon"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all FIM data",
	Long: `Remove all data created by the FIM tool, including:
- Baseline files
- Log files
- PID files
- Configuration files

If the daemon is running, it will be stopped first.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if daemon is running and stop it
		if daemon.IsRunning() {
			fmt.Println("Stopping FIM daemon...")
			if err := stopDaemon(); err != nil {
				return fmt.Errorf("failed to stop daemon: %v", err)
			}
			fmt.Println("FIM daemon stopped successfully")
		}

		// Get home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}

		// Define paths to clean
		fimDir := filepath.Join(homeDir, ".fim")
		configFile := filepath.Join(homeDir, "fim.conf")

		// Remove .fim directory
		if _, err := os.Stat(fimDir); err == nil {
			fmt.Printf("Removing FIM data directory: %s\n", fimDir)
			if err := os.RemoveAll(fimDir); err != nil {
				return fmt.Errorf("failed to remove FIM directory: %v", err)
			}
		}

		// Remove config file
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf("Removing config file: %s\n", configFile)
			if err := os.Remove(configFile); err != nil {
				return fmt.Errorf("failed to remove config file: %v", err)
			}
		}

		fmt.Println("All FIM data has been cleaned successfully")
		return nil
	},
}

// stopDaemon stops the FIM daemon
func stopDaemon() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}
	pidFile := filepath.Join(homeDir, ".fim", "fim.pid")

	// Read PID file
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %v", err)
	}

	// Parse PID
	pid := strings.TrimSpace(string(pidData))
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("failed to parse PID: %v", err)
	}

	// Send SIGTERM to process
	process, err := os.FindProcess(pidInt)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	if err := process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("failed to send signal to process: %v", err)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
