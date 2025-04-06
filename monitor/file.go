package monitor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// FileInfo represents information about a file
type FileInfo struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	Mode      uint32 `json:"mode"`
	ModTime   int64  `json:"mod_time"`
	Hash      string `json:"hash,omitempty"`
	UID       int    `json:"uid"`
	GID       int    `json:"gid"`
	IsDir     bool   `json:"is_dir"`
	IsSymlink bool   `json:"is_symlink"`
}

// ChangeType represents the type of change detected
type ChangeType int

const (
	NoChange ChangeType = iota
	NewFile
	ModifiedFile
	DeletedFile
	PermissionChange
)

// Change represents a detected change in the file system
type Change struct {
	Path      string     `json:"path"`
	Type      ChangeType `json:"type"`
	OldInfo   *FileInfo  `json:"old_info,omitempty"`
	NewInfo   *FileInfo  `json:"new_info,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// GetFileInfo collects information about a file
func GetFileInfo(path string) (*FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	fileInfo := &FileInfo{
		Path:      path,
		Size:      info.Size(),
		Mode:      uint32(info.Mode()),
		ModTime:   info.ModTime().Unix(),
		IsDir:     info.IsDir(),
		IsSymlink: info.Mode()&os.ModeSymlink != 0,
	}

	// Get ownership information
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		fileInfo.UID = int(stat.Uid)
		fileInfo.GID = int(stat.Gid)
	} else {
		// Fallback for systems where Sys() doesn't return *syscall.Stat_t
		// Set default values
		fileInfo.UID = -1
		fileInfo.GID = -1
	}

	// Calculate hash for regular files
	if !info.IsDir() && !fileInfo.IsSymlink {
		hash, err := calculateFileHash(path)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate file hash: %v", err)
		}
		fileInfo.Hash = hash
	}

	// If this is a symlink, also store the target path
	if fileInfo.IsSymlink {
		targetPath, err := filepath.EvalSymlinks(path)
		if err == nil {
			// Add a note about the target in the path
			fileInfo.Path = fmt.Sprintf("%s -> %s", path, targetPath)
		}
	}

	return fileInfo, nil
}

// calculateFileHash calculates the SHA-256 hash of a file
func calculateFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CompareFiles compares two FileInfo objects and returns the type of change
func CompareFiles(oldInfo, newInfo *FileInfo) ChangeType {
	if oldInfo == nil && newInfo != nil {
		return NewFile
	}
	if oldInfo != nil && newInfo == nil {
		return DeletedFile
	}
	if oldInfo == nil || newInfo == nil {
		return NoChange
	}

	if oldInfo.Hash != newInfo.Hash {
		return ModifiedFile
	}

	if oldInfo.Mode != newInfo.Mode || oldInfo.UID != newInfo.UID || oldInfo.GID != newInfo.GID {
		return PermissionChange
	}

	return NoChange
}

// Equals checks if two FileInfo objects are equal
func (f *FileInfo) Equals(other *FileInfo) bool {
	if f == nil || other == nil {
		return false
	}

	// Compare basic properties
	if f.Path != other.Path ||
		f.Size != other.Size ||
		f.Mode != other.Mode ||
		f.ModTime != other.ModTime ||
		f.UID != other.UID ||
		f.GID != other.GID ||
		f.IsDir != other.IsDir ||
		f.IsSymlink != other.IsSymlink {
		return false
	}

	// For regular files, compare hashes
	if !f.IsDir && !f.IsSymlink {
		return f.Hash == other.Hash
	}

	return true
}
