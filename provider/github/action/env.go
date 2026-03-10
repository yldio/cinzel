// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

// EnvConfig represents the HCL configuration for a single env block.
type EnvConfig struct {
	Name  hcl.Expression `hcl:"name,attr"`
	Value hcl.Expression `hcl:"value,attr"`
}

// EnvListConfig is a slice of EnvConfig decoded from HCL env blocks.
type EnvListConfig []EnvConfig

func (config *EnvConfig) parseName(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.Name, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}

func (config *EnvConfig) parseValue(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.Value, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}

// Parse resolves env blocks into a cty object mapping names to values.
func (config *EnvListConfig) Parse(hv *hclparser.HCLVars) (cty.Value, error) {
	if config == nil {
		return cty.NilVal, nil
	}

	mapping := map[string]cty.Value{}

	for _, w := range *config {
		name, err := w.parseName(hv)
		if err != nil {
			return cty.NilVal, err
		}

		if name == cty.NilVal {
			return cty.NilVal, errors.New("name must be set")
		}

		value, err := w.parseValue(hv)
		if err != nil {
			return cty.NilVal, err
		}

		if value != cty.NilVal {
			mapping[name.AsString()] = value
		} else {
			return cty.NilVal, errors.New("value must be set")
		}
	}

	return cty.ObjectVal(mapping), nil
}
