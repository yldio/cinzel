// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package filewriter

import (
	"os"
)

type Writer struct{}

func New() *Writer {
	return &Writer{}
}

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
