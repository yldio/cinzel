// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package yamlparser

import (
	"bytes"
	"errors"

	"github.com/goccy/go-yaml"
	"github.com/yldio/atos/internal/workflow"
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
		// `atos` uses `any` so we need this "hack" to clean `"on":` to just `on:`.
		filteredOut := bytes.Replace(out, []byte("\"on\""), []byte("on"), -1)

		yamls[workflow.Filename] = filteredOut
	}

	return yamls, nil
}
