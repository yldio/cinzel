// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package filewriter

import (
	"path/filepath"
	"testing"
)

func TestWriter(t *testing.T) {
	t.Run("writes to a file", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.yaml")
		content := []byte("abc")

		if err := New().Do(filePath, content); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("create a file fails", func(t *testing.T) {
		filePath := ""
		content := []byte("abc")

		if err := New().Do(filePath, content); err == nil {
			t.Fatal(err)
		}
	})

	t.Run("write to a file fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := []byte("abc")

		filePath := filepath.Join(tmpDir, "dummy-file.yaml")

		if err := New().Do(filePath, content); err != nil {
			t.Fatal(err)
		}
	})
}
