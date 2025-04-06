package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rhinocodelab/IntegrityWatchdog/config"
	"github.com/rhinocodelab/IntegrityWatchdog/monitor"
	"github.com/rhinocodelab/IntegrityWatchdog/storage"
)

// Daemon represents the FIM daemon
type Daemon struct {
	config   *config.Config
	baseline *storage.Baseline
	logger   *log.Logger
	pidFile  string
	running  bool
	interval time.Duration
}

// NewDaemon creates a new daemon instance
func NewDaemon(cfg *config.Config, interval time.Duration) (*Daemon, error) {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create .fim directory if it doesn't exist
	fimDir := filepath.Join(homeDir, ".fim")
	if err := os.MkdirAll(fimDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .fim directory: %v", err)
	}

	// Set up log file
	logFile := filepath.Join(fimDir, "fim.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create logger
	logger := log.New(file, "", log.LstdFlags)

	// Create PID file path
	pidFile := filepath.Join(fimDir, "fim.pid")

	return &Daemon{
		config:   cfg,
		logger:   logger,
		pidFile:  pidFile,
		interval: interval,
	}, nil
}

// Start starts the daemon
func (d *Daemon) Start() error {
	// Check if daemon is already running
	if d.running {
		return fmt.Errorf("daemon is already running")
	}

	// Check if PID file exists
	if _, err := os.Stat(d.pidFile); err == nil {
		return fmt.Errorf("daemon is already running (PID file exists)")
	}

	// Write PID file
	if err := os.WriteFile(d.pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}

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
	d.baseline = baseline

	// Set running flag
	d.running = true

	// Start monitoring loop
	go d.monitorLoop()

	return nil
}

// Stop stops the daemon
func (d *Daemon) Stop() error {
	if !d.running {
		return fmt.Errorf("daemon is not running")
	}

	// Remove PID file
	if err := os.Remove(d.pidFile); err != nil {
		return fmt.Errorf("failed to remove PID file: %v", err)
	}

	// Set running flag
	d.running = false

	return nil
}

// IsRunning checks if the daemon is running
func IsRunning() bool {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	// Check PID file
	pidFile := filepath.Join(homeDir, ".fim", "fim.pid")
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false
	}

	// Read PID
	pidData, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	// Parse PID
	pidStr := string(pidData)
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return process != nil
}

// monitorLoop runs the monitoring loop
func (d *Daemon) monitorLoop() {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for d.running {
		select {
		case <-ticker.C:
			if err := d.scan(); err != nil {
				d.logger.Printf("Error during scan: %v", err)
			}
		}
	}
}

// scan performs a scan and compares with baseline
func (d *Daemon) scan() error {
	// Scan each monitored path
	for _, path := range d.config.Monitor.Paths {
		if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip excluded paths
			if d.config.IsExcluded(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Get current file info
			currentInfo, err := monitor.GetFileInfo(path)
			if err != nil {
				return err
			}

			// Compare with baseline
			baselineInfo, exists := d.baseline.GetFile(path)
			if !exists {
				// New file
				d.logger.Printf("[+] New file: %s", path)
			} else {
				// Check for changes
				if !baselineInfo.Equals(currentInfo) {
					d.logger.Printf("[*] Modified file: %s", path)
				}
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to scan path %s: %v", path, err)
		}
	}

	// Check for deleted files
	for path := range d.baseline.Files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			d.logger.Printf("[-] Deleted file: %s", path)
		}
	}

	return nil
}
