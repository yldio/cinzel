// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

type HclConfig struct {
	Jobs JobsConfig `hcl:"job,block"`
}

func (config *HclConfig) Parse() (Jobs, error) {
	if config.Jobs == nil {
		return Jobs{}, nil
	}

	parsedJobs, err := config.Jobs.Parse()
	if err != nil {
		return Jobs{}, err
	}

	return parsedJobs, nil
}

func Convert(content any) ([]byte, error) {
	out, err := yaml.Marshal(content)
	if err != nil {
		return []byte{}, err
	}

	// Please link to https://github.com/go-yaml/yaml?tab=readme-ov-file#yaml-support-for-the-go-language
	// `atos` uses `any` so we need this "hack" to clean `"on":` to just `on:`.
	filteredOut := bytes.Replace(out, []byte("\"on\""), []byte("on"), -1)

	return filteredOut, nil
}
