// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import "github.com/zclconf/go-cty/cty"

type VariableConfig struct {
	Name  string    `hcl:"name,label"`
	Value cty.Value `hcl:"value,attr"`
}

func (config *VariableConfig) Parse() (any, error) {
	return ParseCtyValue(config.Value, []string{
		cty.String.FriendlyName(),
		cty.Number.FriendlyName(),
		cty.Bool.FriendlyName(),
	})
}
