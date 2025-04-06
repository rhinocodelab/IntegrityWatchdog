package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rhinocodelab/IntegrityWatchdog/daemon"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the FIM daemon",
	Long:  "Stop the FIM daemon if it is running",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if daemon is running
		if !daemon.IsRunning() {
			return fmt.Errorf("daemon is not running")
		}

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
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidData)))
		if err != nil {
			return fmt.Errorf("failed to parse PID: %v", err)
		}

		// Send SIGTERM to process
		process, err := os.FindProcess(pid)
		if err != nil {
			return fmt.Errorf("failed to find process: %v", err)
		}

		if err := process.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed to send signal to process: %v", err)
		}

		fmt.Println("FIM daemon stopped")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
