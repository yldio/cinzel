// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"github.com/zclconf/go-cty/cty"
)

// VariablesConfig is a slice of VariableConfig decoded from HCL variable blocks.
type VariablesConfig []VariableConfig

// VariableConfig represents a single HCL variable declaration with an id and value.
type VariableConfig struct {
	Id    string    `hcl:"id,label"`
	Value cty.Value `hcl:"value,attr"`
}

// Parse registers all variable values into the given HCLVars store.
func (config *VariablesConfig) Parse(hv *HCLVars) error {

	if config == nil {

		return nil
	}

	for _, variable := range *config {
		hv.Add(variable.Id, variable.Value)
	}

	return nil
}
