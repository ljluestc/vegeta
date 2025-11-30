package storage

import (
	"context"
	"io"
	"strings"
)

// StorageType represents different storage backends
type StorageType int

const (
	StorageTypeLocal StorageType = iota
	StorageTypeGCS
	StorageTypeS3
)

// ParseStoragePath determines the storage type and extracts the path
func ParseStoragePath(path string) (StorageType, string) {
	if strings.HasPrefix(path, "gs://") {
		return StorageTypeGCS, strings.TrimPrefix(path, "gs://")
	} else if strings.HasPrefix(path, "s3://") {
		return StorageTypeS3, strings.TrimPrefix(path, "s3://")
	}
	return StorageTypeLocal, path
}

// StorageBackend interface for different storage implementations
type StorageBackend interface {
	OpenReader(ctx context.Context, path string) (io.ReadCloser, error)
	OpenWriter(ctx context.Context, path string) (io.WriteCloser, error)
	Exists(ctx context.Context, path string) (bool, error)
}

// NewStorageBackend creates a new storage backend based on the path
func NewStorageBackend(path string) (StorageBackend, error) {
	storageType, _ := ParseStoragePath(path)
	
	switch storageType {
	case StorageTypeGCS:
		return NewGCSBackend()
	case StorageTypeS3:
		return NewS3Backend()
	default:
		return NewLocalBackend(), nil
	}
}