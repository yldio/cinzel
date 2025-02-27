// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package yamlwriter

import (
	"bytes"
	"errors"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/yldio/acto/provider/github/workflow"
)

// tbd...
type Yaml struct {
	content workflow.Workflows
}

// tbd...
func New(workflows workflow.Workflows) *Yaml {
	return &Yaml{
		content: workflows,
	}
}

// Parses (or Marshal's) the Yaml.Content into Yaml format as a slice of bytes.
func (config *Yaml) Do() (map[string][]byte, error) {
	yamls := make(map[string][]byte)

	for _, workflow := range config.content {
		if workflow.Filename == "" {
			return map[string][]byte{}, errors.New("filename of workflow must be defined")
		}

		out, err := yaml.Marshal(workflow)
		if err != nil {
			return map[string][]byte{}, err
		}

		// Please link to the documentation [here](https://github.com/go-yaml/yaml?tab=readme-ov-file#yaml-support-for-the-go-language)
		// `acto` uses `any` so we need this "hack" to clean `"on":` to just `on:`.
		filteredOut := bytes.Replace(out, []byte("\"on\""), []byte("on"), -1)

		// hack to clean empty structs such as push: {}
		filteredOut = bytes.Replace(filteredOut, []byte(" {}"), []byte(""), -1)

		yamls[strings.Join([]string{workflow.Filename, "yaml"}, ".")] = filteredOut
	}

	return yamls, nil
}
