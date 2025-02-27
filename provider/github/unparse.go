// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package github

import (
	"github.com/yldio/acto/provider"
)

func (p *GitHub) Unparse(opts provider.ProviderOps) error {
	// if cmd.IsSet("file") && cmd.IsSet("directory") {
	// 	return errors.New("`file` or `directory` cannot be set at the same time")
	// } else if !cmd.IsSet("file") && !cmd.IsSet("directory") {
	// 	return errors.New("`file` or `directory` must be set")
	// }

	// var (
	// 	path            string
	// 	outputDirectory string
	// 	recursive       bool = cmd.Bool("recursive")
	// 	dryRun          bool = cmd.Bool("dry-run")
	// )

	// if cmd.IsSet("file") {
	// 	path = cmd.String("file")
	// } else if cmd.IsSet("directory") {
	// 	path = cmd.String("directory")
	// }

	// if cmd.IsSet("output-directory") {
	// 	outputDirectory = cmd.String("output-directory")
	// } else {
	// 	outputDirectory = "./acto"
	// }

	// fileReader := filereader.New()

	// parsedWorkflows, err := fileReader.DoYaml(path, recursive)
	// if err != nil {
	// 	return err
	// }

	// fw := filewriter.New()

	// for _, parsedWorkflow := range parsedWorkflows {
	// 	bytes, err := parsedWorkflow.Decode()
	// 	if err != nil {
	// 		return err
	// 	}

	// 	file := fmt.Sprintf("%s%s", parsedWorkflow.Filename, filereader.FileHclExtensions[0])
	// 	filePath := filepath.Join(outputDirectory, file)

	// 	if dryRun {
	// 		fmt.Printf("# file: %s\n", filePath)
	// 		fmt.Println(string(bytes))
	// 		continue
	// 	}

	// 	if err := fw.Do(filePath, bytes); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}
