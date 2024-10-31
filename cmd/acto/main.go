// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

// Package acto (pronounced as "AH-toosh" (IPA: /ˈa.tuʃ/))
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/service/actoflag"
	"github.com/yldio/acto/service/filereader"
	"github.com/yldio/acto/service/filewriter"
	"github.com/yldio/acto/service/hclparser"
)

var (
	version = ""
)

const gitHubDir = ".github/workflows"

func do(flags *actoflag.Flags, outputDir string) error {
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

	actoReader := filereader.New(path, flags.Recursive)
	actoReader.ReadDir(filereader.FileHCLExt)
	fmt.Println(actoReader.GetFiles())

	bodies, err := actoReader.Do(filereader.FileHCLExt)
	if err != nil {
		return err
	}

	body := hcl.MergeBodies(bodies)

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
		actoWriter := filewriter.New()
		filePath := fmt.Sprintf("%s/%s", outputDir, file)
		if err := actoWriter.Do(filePath, content); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	cliApp := actoflag.NewCliApp()

	if err := cliApp.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	// flags := actoflag.New()

	// // TODO: move
	// if err := os.MkdirAll(gitHubDir, os.ModePerm); err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }

	// if err := do(flags, gitHubDir); err != nil {
	// 	fmt.Println(err.Error())
	// 	os.Exit(1)
	// }
}
