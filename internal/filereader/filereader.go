// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/acto/internal/actoerrors"
)

var (
	FileHclExtensions  = []string{".hcl"}
	FileYamlExtensions = []string{".yaml", ".yml"}
)

type Parser interface{}

type Reader struct {
	files []string
}

func (read *Reader) GetFiles() []string {
	return read.files
}

func New() *Reader {
	return &Reader{}
}

func (read *Reader) readPath(path string, recursive bool, extensions []string) error {
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

func (read *Reader) FromHCL(path string, recursive bool) (hcl.Body, error) {
	if err := read.readPath(path, recursive, FileHclExtensions); err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()
	var bodies []hcl.Body

	for _, hclFile := range read.files {
		file, diags := parser.ParseHCLFile(hclFile)
		if diags.HasErrors() {
			return nil, actoerrors.ProcessHCLDiags(diags)
		}

		bodies = append(bodies, file.Body)
	}

	return hcl.MergeBodies(bodies), nil
}
