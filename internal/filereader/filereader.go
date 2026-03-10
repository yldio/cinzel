// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package filereader

import (
	"os"
	"path/filepath"
	"slices"
)

// Updater is implemented by types that can receive a filename after being read.
type Updater interface {
	Update(string)
}

// Reader discovers and reads files of a given type T from disk.
type Reader[T Updater] struct {
	files []string
}

// New returns a new Reader instance.
func New[T Updater]() *Reader[T] {
	return &Reader[T]{}
}

// GetFiles returns the list of file paths discovered by the reader.
func (read *Reader[T]) GetFiles() []string {
	return read.files
}

func (read *Reader[T]) readPath(path string, recursive bool, extensions []string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		if !slices.Contains(extensions, filepath.Ext(path)) {
			return nil
		}

		read.files = append(read.files, path)

		return nil
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullpath := filepath.Join(path, file.Name())

		info, err := os.Stat(fullpath)
		if err != nil {
			return err
		}

		if !recursive && info.IsDir() {
			continue
		}

		if err := read.readPath(fullpath, recursive, extensions); err != nil {
			return err
		}
	}

	return nil
}
