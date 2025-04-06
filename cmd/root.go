package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	daemonMode bool
)

var rootCmd = &cobra.Command{
	Use:   "fim",
	Short: "File Integrity Monitoring Tool",
	Long: `A lightweight file integrity monitoring CLI tool for Linux/Unix systems.
It helps detect unauthorized or unexpected changes in your file system.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&daemonMode, "daemon", "d", false, "Run in daemon mode")
}
