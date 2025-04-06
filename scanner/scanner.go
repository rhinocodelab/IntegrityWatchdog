package scanner

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rhinocodelab/IntegrityWatchdog/config"
	"github.com/rhinocodelab/IntegrityWatchdog/monitor"
	"github.com/rhinocodelab/IntegrityWatchdog/storage"
)

// Scanner represents a file system scanner
type Scanner struct {
	config *config.Config
}

// NewScanner creates a new scanner with the given configuration
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config: cfg,
	}
}

// ScanPaths scans all configured paths and returns a baseline
func (s *Scanner) ScanPaths() (*storage.Baseline, error) {
	baseline := storage.NewBaseline()

	// Scan each monitored path
	for _, path := range s.config.Monitor.Paths {
		// Check if the path is a symlink
		realPath := path
		if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
			// Resolve the symlink
			var err error
			realPath, err = filepath.EvalSymlinks(path)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve symlink %s: %v", path, err)
			}
		}

		if err := filepath.Walk(realPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip excluded paths
			if s.config.IsExcluded(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Collect file information
			fileInfo, err := monitor.GetFileInfo(path)
			if err != nil {
				return err
			}

			baseline.AddFile(fileInfo)
			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to scan path %s: %v", path, err)
		}
	}

	return baseline, nil
}
