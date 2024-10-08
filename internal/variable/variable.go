// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package variable

import (
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type VariablesConfig []VariableConfig

type VariableConfig struct {
	Id    string    `hcl:"id,label"`
	Value cty.Value `hcl:"value,attr"`
}

func (config *VariablesConfig) Parse() error {
	for _, variable := range *config {
		value, err := actoparser.ParseCtyValue(variable.Value, []string{
			cty.String.FriendlyName(),
			cty.Number.FriendlyName(),
			cty.Bool.FriendlyName(),
			cty.EmptyTuple.FriendlyName(),
			cty.DynamicPseudoType.FriendlyName(),
		})
		if err != nil {
			return err
		}

		vars := variables.Instance()

		vars.Add(variable.Id, value)
	}

	return nil
}
