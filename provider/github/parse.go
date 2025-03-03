// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package github

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/yldio/acto/internal/filereader"
	"github.com/yldio/acto/internal/filewriter"
	"github.com/yldio/acto/internal/yamlwriter"
	"github.com/yldio/acto/provider"
)

type hclBody struct{}

func (h *hclBody) Update(filename string) {}

func (p *GitHub) Parse(opts provider.ProviderOps) error {
	if opts.File == "" && opts.Directory == "" {
		return errors.New("`file` or `directory` cannot be set at the same time")
	} else if opts.File != "" && opts.Directory != "" {
		return errors.New("`file` or `directory` must be set")
	}

	var (
		path            string
		outputDirectory string
		recursive       bool = opts.Recursive
		dryRun          bool = opts.DryRun
	)

	if opts.File != "" {
		path = opts.File
	} else if opts.Directory != "" {
		path = opts.Directory
	}

	if opts.OutputDirectory != "" {
		outputDirectory = opts.OutputDirectory
	} else {
		outputDirectory = p.DefaultOutputDirectory()
	}

	fileReader := filereader.Reader[*hclBody]{}

	hclBody, err := fileReader.FromHCL(path, recursive)
	if err != nil {
		return err
	}

	p.bodyHCL = hclBody

	parsedWorkflows, err := p.Do()
	if err != nil {
		return err
	}

	yamlWriter := yamlwriter.New(parsedWorkflows)

	listOfFiles, err := yamlWriter.Do()
	if err != nil {
		return err
	}

	fw := filewriter.New()

	for file, bytes := range listOfFiles {
		filePath := filepath.Join(outputDirectory, file)

		if dryRun {
			fmt.Printf("# file: %s\n", filePath)
			fmt.Println(string(bytes))
			continue
		}

		if err := fw.Do(filePath, bytes); err != nil {
			return err
		}
	}

	return nil
}
