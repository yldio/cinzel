// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/cinzel/internal/cinzelerror"
)

// ErrNoHCLFiles is returned when no HCL files are found in the given path.
var ErrNoHCLFiles = errors.New("no HCL files found in input")

// ParseHCLInput parses one or more HCL files from path and returns a merged body.
func ParseHCLInput(path string, recursive bool) (hcl.Body, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !stat.IsDir() {
		parser := hclparse.NewParser()
		file, diags := parser.ParseHCLFile(path)

		if diags.HasErrors() {
			return nil, cinzelerror.ProcessHCLDiags(diags)
		}

		return file.Body, nil
	}

	var bodies []hcl.Body
	parser := hclparse.NewParser()

	err = filepath.WalkDir(path, func(current string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if current != path && !recursive {
				return filepath.SkipDir
			}

			return nil
		}

		if filepath.Ext(current) != ".hcl" {
			return nil
		}

		file, diags := parser.ParseHCLFile(current)

		if diags.HasErrors() {
			return cinzelerror.ProcessHCLDiags(diags)
		}

		bodies = append(bodies, file.Body)

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(bodies) == 0 {
		return nil, ErrNoHCLFiles
	}

	return hcl.MergeBodies(bodies), nil
}

// ListFilesWithExtensions returns files under path matching the given extensions.
func ListFilesWithExtensions(path string, recursive bool, exts ...string) ([]string, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	allowed := map[string]struct{}{}

	for _, ext := range exts {
		allowed[strings.ToLower(ext)] = struct{}{}
	}

	if !stat.IsDir() {
		ext := strings.ToLower(filepath.Ext(path))

		if _, ok := allowed[ext]; !ok {
			return nil, fmt.Errorf("input file must have %s extension", joinExtensions(exts))
		}

		return []string{path}, nil
	}

	files := []string{}
	err = filepath.WalkDir(path, func(current string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if current != path && !recursive {
				return filepath.SkipDir
			}

			return nil
		}

		ext := strings.ToLower(filepath.Ext(current))

		if _, ok := allowed[ext]; ok {
			files = append(files, current)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// WriteFile writes content to path, creating parent directories as needed.
func WriteFile(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0o644)
}

func joinExtensions(exts []string) string {
	if len(exts) == 0 {
		return "a supported"
	}

	if len(exts) == 1 {
		return exts[0]
	}

	if len(exts) == 2 {
		return exts[0] + " or " + exts[1]
	}

	return strings.Join(exts[:len(exts)-1], ", ") + ", or " + exts[len(exts)-1]
}
