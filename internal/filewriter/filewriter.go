// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package filewriter

import (
	"os"
)

// Writer writes byte content to files on disk.
type Writer struct{}

// New returns a new Writer instance.
func New() *Writer {

	return &Writer{}
}

// Do creates or truncates filePath and writes content to it.
func (writer *Writer) Do(filePath string, content []byte) error {
	file, err := os.Create(filePath)
	if err != nil {

		return err
	}

	defer file.Close()

	_, err = file.Write(content)
	if err != nil {

		return err
	}

	return nil
}
