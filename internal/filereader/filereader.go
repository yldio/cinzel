// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"os"
	"path/filepath"
	"slices"
)

type Updater interface {
	Update(string)
}

type Reader[T Updater] struct {
	files []string
}

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
