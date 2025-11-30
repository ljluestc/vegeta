package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	vegeta "github.com/tsenart/vegeta/v12/lib"
	"github.com/tsenart/vegeta/v12/lib/storage"
)

// fileWithStorage opens a file or cloud storage object for reading/writing
func fileWithStorage(name string, create bool) (io.Closer, error) {
	ctx := context.Background()
	
	// Handle stdin/stdout
	switch name {
	case "stdin":
		return os.Stdin, nil
	case "stdout":
		return os.Stdout, nil
	}
	
	// Check if it's a cloud storage path
	storageType, storagePath := storage.ParseStoragePath(name)
	if storageType != storage.StorageTypeLocal {
		backend, err := storage.NewStorageBackend(name)
		if err != nil {
			return nil, fmt.Errorf("failed to create storage backend for %s: %w", name, err)
		}
		
		if create {
			return backend.OpenWriter(ctx, storagePath)
		} else {
			return backend.OpenReader(ctx, storagePath)
		}
	}
	
	// Handle local files
	if create {
		return os.Create(name)
	}
	return os.Open(name)
}

// fileReader opens a file or cloud storage object for reading
func fileReader(name string) (io.ReadCloser, error) {
	closer, err := fileWithStorage(name, false)
	if err != nil {
		return nil, err
	}
	
	// Type assertion to ensure it's a ReadCloser
	if reader, ok := closer.(io.ReadCloser); ok {
		return reader, nil
	}
	
	return nil, fmt.Errorf("fileReader: returned closer is not a ReadCloser")
}

// fileWriter opens a file or cloud storage object for writing
func fileWriter(name string) (io.WriteCloser, error) {
	closer, err := fileWithStorage(name, true)
	if err != nil {
		return nil, err
	}
	
	// Type assertion to ensure it's a WriteCloser
	if writer, ok := closer.(io.WriteCloser); ok {
		return writer, nil
	}
	
	return nil, fmt.Errorf("fileWriter: returned closer is not a WriteCloser")
}

func file(name string, create bool) (*os.File, error) {
	switch name {
	case "stdin":
		return os.Stdin, nil
	case "stdout":
		return os.Stdout, nil
	default:
		if create {
			return os.Create(name)
		}
		return os.Open(name)
	}
}

func decoder(files []string) (vegeta.Decoder, io.Closer, error) {
	closer := make(multiCloser, 0, len(files))
	decs := make([]vegeta.Decoder, 0, len(files))
	for _, f := range files {
		rc, err := fileReader(f)
		if err != nil {
			return nil, closer, err
		}

		dec := vegeta.DecoderFor(rc)
		if dec == nil {
			return nil, closer, fmt.Errorf("encode: can't detect encoding of %q", f)
		}

		decs = append(decs, dec)
		closer = append(closer, rc)
	}
	return vegeta.NewRoundRobinDecoder(decs...), closer, nil
}

type multiCloser []io.Closer

func (mc multiCloser) Close() error {
	var errs []string
	for _, c := range mc {
		if err := c.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}
