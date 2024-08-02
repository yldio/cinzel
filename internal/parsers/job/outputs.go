// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"errors"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type OutputConfig struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type OutputsConfig []OutputConfig

type Outputs map[string]any

func (config *OutputConfig) Parse() (string, any, error) {
	switch config.Value.Type().FriendlyName() {
	case cty.String.FriendlyName():
		var val string
		err := gocty.FromCtyValue(config.Value, &val)
		if err != nil {
			return "", nil, err
		}
		return config.Name, val, nil
	case cty.Number.FriendlyName():
		var val int32
		err := gocty.FromCtyValue(config.Value, &val)
		if err != nil {
			var val float32
			err := gocty.FromCtyValue(config.Value, &val)
			if err != nil {
				return "", nil, err
			}
		}
		return config.Name, val, nil
	case cty.Bool.FriendlyName():
		var val bool
		err := gocty.FromCtyValue(config.Value, &val)
		if err != nil {
			return "", nil, err
		}
		return config.Name, val, nil
	default:
		return "", nil, errors.New("invalid output")
	}
}

func (config *OutputsConfig) Parse() (Outputs, error) {
	outputs := make(Outputs)

	for _, output := range *config {
		key, val, err := output.Parse()
		if err != nil {
			return Outputs{}, err
		}

		outputs[key] = val
	}

	return outputs, nil
}
