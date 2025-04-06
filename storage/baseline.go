package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rhinocodelab/IntegrityWatchdog/monitor"
)

// Baseline represents the stored state of monitored files
type Baseline struct {
	Files     map[string]*monitor.FileInfo `json:"files"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
	mu        sync.RWMutex
}

// Changes represents the differences between two baselines
type Changes struct {
	Added    []*monitor.FileInfo `json:"added"`
	Modified []*monitor.FileInfo `json:"modified"`
	Deleted  []*monitor.FileInfo `json:"deleted"`
}

// NewBaseline creates a new baseline
func NewBaseline() *Baseline {
	return &Baseline{
		Files:     make(map[string]*monitor.FileInfo),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddFile adds or updates a file in the baseline
func (b *Baseline) AddFile(info *monitor.FileInfo) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.Files[info.Path] = info
	b.UpdatedAt = time.Now()
}

// GetFile retrieves file information from the baseline
func (b *Baseline) GetFile(path string) (*monitor.FileInfo, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	info, exists := b.Files[path]
	return info, exists
}

// RemoveFile removes a file from the baseline
func (b *Baseline) RemoveFile(path string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.Files, path)
	b.UpdatedAt = time.Now()
}

// Save saves the baseline to a JSON file
func (b *Baseline) Save(filepath string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// Load loads a baseline from a JSON file
func Load(filepath string) (*Baseline, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}

	return &baseline, nil
}

// GetDefaultBaselinePath returns the default path for storing the baseline file
func GetDefaultBaselinePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "fim-baseline.json"
	}
	return filepath.Join(homeDir, ".fim", "baseline.json")
}

// Compare compares this baseline with another baseline and returns the changes
func (b *Baseline) Compare(other *Baseline) *Changes {
	changes := &Changes{
		Added:    make([]*monitor.FileInfo, 0),
		Modified: make([]*monitor.FileInfo, 0),
		Deleted:  make([]*monitor.FileInfo, 0),
	}

	// Check for added and modified files
	for path, otherFile := range other.Files {
		baselineFile, exists := b.Files[path]
		if !exists {
			// File is new
			changes.Added = append(changes.Added, otherFile)
		} else {
			// Check if file is modified
			if !baselineFile.Equals(otherFile) {
				changes.Modified = append(changes.Modified, otherFile)
			}
		}
	}

	// Check for deleted files
	for path, baselineFile := range b.Files {
		if _, exists := other.Files[path]; !exists {
			// File is deleted
			changes.Deleted = append(changes.Deleted, baselineFile)
		}
	}

	return changes
}
