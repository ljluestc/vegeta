package storage

import (
	"context"
	"io"
	"os"
)

// LocalBackend implements StorageBackend for local filesystem
type LocalBackend struct{}

// NewLocalBackend creates a new LocalBackend
func NewLocalBackend() *LocalBackend {
	return &LocalBackend{}
}

// OpenReader opens a file for reading
func (lb *LocalBackend) OpenReader(ctx context.Context, path string) (io.ReadCloser, error) {
	if path == "stdin" {
		return os.Stdin, nil
	}
	return os.Open(path)
}

// OpenWriter opens a file for writing
func (lb *LocalBackend) OpenWriter(ctx context.Context, path string) (io.WriteCloser, error) {
	if path == "stdout" {
		return os.Stdout, nil
	}
	return os.Create(path)
}

// Exists checks if a file exists
func (lb *LocalBackend) Exists(ctx context.Context, path string) (bool, error) {
	if path == "stdin" || path == "stdout" {
		return true, nil
	}
	
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
