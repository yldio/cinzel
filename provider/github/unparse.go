// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package github

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/yldio/acto/internal/filereader"
	"github.com/yldio/acto/internal/filewriter"
	"github.com/yldio/acto/provider"
	"github.com/yldio/acto/provider/github/workflow"
)

func (p *GitHub) Unparse(opts provider.ProviderOps) error {
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

	fileReader := filereader.Reader[*workflow.Workflow]{}

	parsedWorkflows, err := fileReader.FromYaml(path, recursive)
	if err != nil {
		return err
	}

	fw := filewriter.New()

	for _, parsedWorkflow := range parsedWorkflows {
		bytes, err := parsedWorkflow.Decode()
		if err != nil {
			return err
		}

		file := fmt.Sprintf("%s%s", parsedWorkflow.Filename, ".hcl")
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
