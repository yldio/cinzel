// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"github.com/zclconf/go-cty/cty"
)

type OutputConfig struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type OutputsConfig []*OutputConfig

func (config *OutputConfig) Parse() (any, error) {
	if config == nil {
		return nil, nil
	}

	return ParseCtyValue(config.Value, []string{
		cty.String.FriendlyName(),
		cty.Number.FriendlyName(),
		cty.Bool.FriendlyName(),
		cty.EmptyTuple.FriendlyName(),
	})
}

func (config *OutputsConfig) Parse() (map[string]any, error) {
	if config == nil || len(*config) == 0 {
		return nil, nil
	}

	outputs := make(map[string]any)

	for _, output := range *config {
		val, err := output.Parse()
		if err != nil {
			return map[string]any{}, err
		}

		outputs[output.Name] = val
	}

	return outputs, nil
}
