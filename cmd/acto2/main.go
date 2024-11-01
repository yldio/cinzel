// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

// Package acto (pronounced as "AH-toosh" (IPA: /ˈa.tuʃ/))
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/yldio/acto/internal/workflow"
	"github.com/yldio/acto/service/filewriter"
)

func main() {
	path := "../../.github/workflows"

	files, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())

		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		fp := filepath.Join(path, file.Name())

		b, err := os.ReadFile(fp)
		if err != nil {
		}

		var w workflow.Workflow

		yaml.Unmarshal(b, &w)

		filename := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		bytes, err := w.Decode(filename)
		if err != nil {
			panic(err)
		}

		fw := filewriter.New()

		fw.Do(fmt.Sprintf("%s.hcl", filename), bytes)
	}
}
