// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

// Package acto (pronounced as "AH-toosh" (IPA: /ˈa.tuʃ/))
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/yldio/acto/service/actoflag"
	"github.com/yldio/acto/service/filereader"
	"github.com/yldio/acto/service/hclparser"
	"github.com/yldio/acto/service/yamlwriter"
)

var (
	version = ""
)

func getPath(flags *actoflag.ActoCli) (string, error) {
	if flags.InputFile != "" {
		return flags.InputFile, nil
	} else if flags.InputDirectory != "" {
		return flags.InputDirectory, nil
	}

	return "", errors.New("no input file or directory defined")
}

func parse(flags *actoflag.ActoCli) error {
	path, err := getPath(flags)
	if err != nil {
		return err
	}

	fileReader := filereader.New()

	hclBody, err := fileReader.DoHcl(path, flags.Recursive)
	if err != nil {
		return err
	}

	hclParser := hclparser.New(hclBody)

	parsedWorkflows, err := hclParser.Do()
	if err != nil {
		return err
	}

	yamlwriter := yamlwriter.New(parsedWorkflows)

	listOfFiles, err := yamlwriter.Do()
	if err != nil {
		return err
	}

	for file, yamlContent := range listOfFiles {
		if flags.DryRun {
			fmt.Printf("# file: %s\n", file)
			fmt.Println(string(yamlContent))
			continue
		}
	}

	return nil
}

func decode(flags *actoflag.ActoCli) error {
	path, err := getPath(flags)
	if err != nil {
		return err
	}

	fileReader := filereader.New()
	parsedWorkflows, err := fileReader.DoYaml(path, flags.Recursive)
	if err != nil {
		return err
	}

	for _, parsedWorkflow := range parsedWorkflows {
		bytes, err := parsedWorkflow.Decode()
		if err != nil {
			return err
		}

		if flags.DryRun {
			fmt.Printf("# file: %s\n", parsedWorkflow.Filename)
			fmt.Println(string(bytes))
			continue
		}

		// fw := filewriter.New()
		// fw.Do(fmt.Sprintf("%s.hcl", parsedWorkflow.Filename), bytes)
	}

	return nil
}

func main() {
	flags := actoflag.NewFlags()

	cliApp := actoflag.NewCliApp(flags)

	if err := cliApp.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var err error

	if !flags.FromActions {
		err = parse(flags)
	} else {
		err = decode(flags)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
