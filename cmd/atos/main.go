// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

// Package atos (pronounced as "AH-toosh" (IPA: /ˈa.tuʃ/))
package main

import (
	"fmt"
	"os"

	"github.com/yldio/atos/service/flag"
	"github.com/yldio/atos/service/hclparser"
	"github.com/yldio/atos/service/reader"
	"github.com/yldio/atos/service/writer"
)

func do(flags *flag.Flags, outputDir string) error {
	atosReader := reader.New(flags.Directory, flags.File, flags.Recursive)

	bodies, err := atosReader.Do()
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

		fileInfo, err := os.Stat(outputDir)
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			return fmt.Errorf("%s is not a directory", outputDir)
		}

		for file, content := range listOfFiles {
			atosWriter := writer.New()
			filePath := fmt.Sprintf("%s/%s", outputDir, file)
			if err := atosWriter.Do(filePath, content); err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	flags := flag.NewParseFlags()
	gitHubDir := "./github/workflows"

	if err := do(flags, gitHubDir); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
