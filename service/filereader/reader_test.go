// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/service/filewriter"
)

func TestReader(t *testing.T) {
	t.Run("reads from a file", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.hcl")
		content := []byte("workflow \"workflow1\" {}")

		if err := filewriter.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		actoReader := New(filePath, false)

		_, err := actoReader.Do()
		if err != nil {
			t.Fatal(err.Error())
		}
	})

	t.Run("reads from a directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.hcl")
		content := []byte("workflow \"workflow1\" {}")

		if err := filewriter.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		actoReader := New(tmpDir, false)

		_, err := actoReader.Do()
		if err != nil {
			t.Fatal(err.Error())
		}
	})

	t.Run("only allow reading from HCL file(s)", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.yaml")
		content := []byte("abc")

		if err := filewriter.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		actoReader := New(filePath, false)

		_, err := actoReader.Do()
		if err == nil {
			t.Fatal("should fail because it's not an HCL file")
		}

		if !errors.Is(err, actoerrors.ErrOnlyHclFiles) {
			t.Fatal("got wrong error message")
		}
	})

	t.Run("should read an HCL file with valid HCL syntax", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.hcl")
		content := []byte("abc")

		if err := filewriter.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		actoReader := New(filePath, false)

		_, err := actoReader.Do()
		if err == nil {
			t.Fatal("should fail because it's not an HCL syntax")
		}
	})
}
