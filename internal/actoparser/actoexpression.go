// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"github.com/hashicorp/hcl/v2"
)

type ActoExpression struct {
	expression hcl.Expression
}

func NewActoExpression(expression hcl.Expression) *ActoExpression {
	return &ActoExpression{
		expression: expression,
	}
}

func (acto *ActoExpression) Parse() (any, error) {
	switch actoType := acto.expression.(type) {
	default:
		val, _ := actoType.Value(nil)
		// val, diags := actoType.Value(nil)
		// if diags.HasErrors() {}

		return val.AsValueMap(), nil
	}
}
