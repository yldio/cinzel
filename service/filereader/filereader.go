// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/workflow"
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

func (read *Reader) DoHcl(path string, recursive bool) (hcl.Body, error) {
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

func (read *Reader) DoYaml(path string, recursive bool) (workflow.Workflows, error) {
	if err := read.readPath(path, recursive, FileYamlExtensions); err != nil {
		return workflow.Workflows{}, err
	}

	var workflows workflow.Workflows

	for _, yamlFile := range read.files {
		yamlBytes, err := os.ReadFile(yamlFile)
		if err != nil {
			return workflow.Workflows{}, err
		}

		var w workflow.Workflow

		yaml.Unmarshal(yamlBytes, &w)

		fileBase := filepath.Base(yamlFile)

		filename := strings.TrimSuffix(fileBase, filepath.Ext(fileBase))

		w.Id = filename
		w.Filename = filename

		workflows = append(workflows, &w)
	}

	return workflows, nil
}
