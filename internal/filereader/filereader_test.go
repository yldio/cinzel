// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package filereader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/internal/test"
)

type tempFile struct {
	filename string
	content  []byte
}

func TempFiles(t *testing.T, tempFiles ...tempFile) []string {
	t.Helper()

	if len(tempFiles) == 0 {
		t.Fatal()
	}

	tmpDir := t.TempDir()
	var filepaths []string

	for _, tempFile := range tempFiles {
		filePath := filepath.Join(tmpDir, tempFile.filename)
		err := os.WriteFile(filePath, tempFile.content, 0644)
		if err != nil {
			t.Fatal(err)
		}

		filepaths = append(filepaths, filePath)
	}

	return filepaths
}

func TestFilereader(t *testing.T) {
	t.Run("reads without errors from an HCL file", func(t *testing.T) {
		content := []byte("workflow \"workflow1\" {}")
		filePath := TempFiles(t, tempFile{"dummy-file.hcl", content})

		fileReader := New[test.HclBody]()

		hclBody, err := fileReader.FromHCL(filePath[0], false)
		if err != nil {
			t.Fatal(err.Error())
		}

		if hclBody == nil {
			t.Fatal()
		}
	})

	t.Run("reads with errors from an HCL file", func(t *testing.T) {
		content := []byte("workflow \"workflow1\"")
		filePath := TempFiles(t, tempFile{"dummy-file.hcl", content})

		message := "Either a quoted string block label or an opening brace (\"{\") is expected here., if you think this is incorrect, consider opening an issue in https://www.github.com/yldio/cinzel/issues"

		fileReader := New[test.HclBody]()

		hclBody, err := fileReader.FromHCL(filePath[0], false)
		if err.Error() != message {
			t.Fatal(err.Error())
		}

		if hclBody != nil {
			t.Fatal()
		}
	})

	t.Run("can't read non existing HCL file", func(t *testing.T) {
		filePath := "somewhere"

		message := "stat somewhere: no such file or directory"

		fileReader := New[test.HclBody]()

		hclBody, err := fileReader.FromHCL(filePath, false)

		if err.Error() != message {
			t.Fatal(err.Error())
		}

		if hclBody != nil {
			t.Fatal()
		}
	})

	t.Run("reads without errors from an YAML file", func(t *testing.T) {
		content := []byte(`name: Pull Request`)
		filePath := TempFiles(t, tempFile{"dummy-file.yaml", content})

		fileReader := New[test.YamlBody]()

		yamlBody, err := fileReader.FromYaml(filePath[0], false)
		if err != nil {
			t.Fatal(err.Error())
		}

		if yamlBody == nil {
			t.Fatal()
		}
	})

	t.Run("reads with errors from an YAML file", func(t *testing.T) {
		content := []byte(`on`)
		filePath := TempFiles(t, tempFile{"dummy-file.yaml", content})

		message := "[1:1] string was used where mapping is expected\n>  1 | on\n       ^\n"

		fileReader := New[test.YamlBody]()

		yamlBody, err := fileReader.FromYaml(filePath[0], false)
		if err.Error() != message {
			t.Fatal(err.Error())
		}

		if yamlBody != nil {
			t.Fatal()
		}
	})

	t.Run("can't read non existing YAML file", func(t *testing.T) {
		filePath := "somewhere"

		message := "stat somewhere: no such file or directory"

		fileReader := New[test.YamlBody]()

		yamlBody, err := fileReader.FromYaml(filePath, false)

		if err.Error() != message {
			t.Fatal(err.Error())
		}

		if yamlBody != nil {
			t.Fatal()
		}
	})
}
