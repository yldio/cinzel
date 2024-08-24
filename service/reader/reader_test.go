// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package reader

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/yldio/atos/service/atoserrors"
	"github.com/yldio/atos/service/writer"
)

func TestReader(t *testing.T) {
	t.Run("reads from a file", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.hcl")
		content := []byte("workflow \"workflow_1\" {}")

		if err := writer.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		atosReader := New(filePath, false)

		_, err := atosReader.Do()
		if err != nil {
			t.Fatal(err.Error())
		}
	})

	t.Run("reads from a directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.hcl")
		content := []byte("workflow \"workflow_1\" {}")

		if err := writer.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		atosReader := New(tmpDir, false)

		_, err := atosReader.Do()
		if err != nil {
			t.Fatal(err.Error())
		}
	})

	t.Run("only allow reading from HCL file(s)", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.yaml")
		content := []byte("abc")

		if err := writer.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		atosReader := New(filePath, false)

		_, err := atosReader.Do()
		if err == nil {
			t.Fatal("should fail because it's not an HCL file")
		}

		if !errors.Is(err, atoserrors.ErrOnlyHclFiles) {
			t.Fatal("got wrong error message")
		}
	})

	t.Run("should read an HCL file with valid HCL syntax", func(t *testing.T) {
		tmpDir := t.TempDir()

		filePath := filepath.Join(tmpDir, "dummy-file.hcl")
		content := []byte("abc")

		if err := writer.New().Do(filePath, content); err != nil {
			t.Fatal(err.Error())
		}

		atosReader := New(filePath, false)

		_, err := atosReader.Do()
		if err == nil {
			t.Fatal("should fail because it's not an HCL syntax")
		}
	})
}
