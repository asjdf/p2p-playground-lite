package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/asjdf/p2p-playground-lite/pkg/types"
)

// FileStorage implements filesystem-based storage
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new filesystem storage
func NewFileStorage(baseDir string) (*FileStorage, error) {
	// Expand home directory
	if strings.HasPrefix(baseDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home dir: %w", err)
		}
		baseDir = filepath.Join(home, baseDir[2:])
	}

	// Create base directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base dir: %w", err)
	}

	return &FileStorage{baseDir: baseDir}, nil
}

// Save stores data under a key
func (s *FileStorage) Save(ctx context.Context, key string, data []byte) error {
	path := filepath.Join(s.baseDir, key)

	// Create parent directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create dir: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return types.WrapError(err, "failed to write file")
	}

	return nil
}

// Load retrieves data by key
func (s *FileStorage) Load(ctx context.Context, key string) ([]byte, error) {
	path := filepath.Join(s.baseDir, key)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, types.ErrNotFound
		}
		return nil, types.WrapError(err, "failed to read file")
	}

	return data, nil
}

// Delete removes data by key
func (s *FileStorage) Delete(ctx context.Context, key string) error {
	path := filepath.Join(s.baseDir, key)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return types.ErrNotFound
		}
		return types.WrapError(err, "failed to delete file")
	}

	return nil
}

// List returns all keys with the given prefix
func (s *FileStorage) List(ctx context.Context, prefix string) ([]string, error) {
	searchPath := filepath.Join(s.baseDir, prefix)
	var keys []string

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore errors for non-existent paths
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if !info.IsDir() {
			// Get relative path from base dir
			relPath, err := filepath.Rel(s.baseDir, path)
			if err != nil {
				return err
			}
			keys = append(keys, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, types.WrapError(err, "failed to list files")
	}

	return keys, nil
}

// Exists checks if a key exists
func (s *FileStorage) Exists(ctx context.Context, key string) (bool, error) {
	path := filepath.Join(s.baseDir, key)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, types.WrapError(err, "failed to check file")
	}

	return true, nil
}

// GetPath returns the absolute path for a key
func (s *FileStorage) GetPath(key string) string {
	return filepath.Join(s.baseDir, key)
}

// BaseDir returns the base directory
func (s *FileStorage) BaseDir() string {
	return s.baseDir
}

// CreateFile creates a new file for writing
func (s *FileStorage) CreateFile(path string) (*os.File, error) {
	// Create parent directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dir: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}
