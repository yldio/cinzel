// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

// Package acto (pronounced as "AH-toosh" (IPA: /ˈa.tuʃ/))
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yldio/acto/service/flag"
	"github.com/yldio/acto/service/hclparser"
	"github.com/yldio/acto/service/reader"
	"github.com/yldio/acto/service/writer"
)

var (
	version = ""
)

func do(flags *flag.Flags, outputDir string) error {
	var path string

	if flags.Version {
		fmt.Printf("Acto version: %s.", version)
		fmt.Println("\nAny feature or issue please read the README on the official GitHub repo.")
		os.Exit(0)
	} else if flags.Help {
		flags.GetUsage()
		os.Exit(0)
	} else if flags.Directory != "" {
		path = flags.Directory
	} else if flags.File != "" {
		path = flags.File
	} else {
		flags.GetUsage()
		os.Exit(0)
	}

	actoReader := reader.New(path, flags.Recursive)

	bodies, err := actoReader.Do()
	if err != nil {
		return err
	}

	for _, body := range bodies {
		parser := hclparser.New(body)

		if err := parser.Decode(); err != nil {
			return err
		}

		content, err := parser.Parse()
		if err != nil {
			return err
		}

		listOfFiles, err := content.Do()
		if err != nil {
			return err
		}

		curDir, err := os.Getwd()
		if err != nil {
			return err
		}

		var fileinfo string

		if filepath.IsAbs(outputDir) {
			fileinfo = outputDir
		} else {
			fileinfo = fmt.Sprintf("%s/%s", curDir, outputDir)
		}

		fileInfo, err := os.Stat(fileinfo)
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			return fmt.Errorf("%s is not a directory", outputDir)
		}

		for file, content := range listOfFiles {
			actoWriter := writer.New()
			filePath := fmt.Sprintf("%s/%s", outputDir, file)
			if err := actoWriter.Do(filePath, content); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	flags := flag.New()
	gitHubDir := ".github/workflows"

	if err := do(flags, gitHubDir); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
