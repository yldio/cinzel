// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
)

// Expression wraps a generic HCL expression for direct evaluation.
type Expression struct {
	expression hcl.Expression
}

// NewExpression creates an Expression wrapper for the given HCL expression.
func NewExpression(expression hcl.Expression) *Expression {

	return &Expression{
		expression: expression,
	}
}

// Parse evaluates the expression and returns its cty value.
func (e *Expression) Parse() (cty.Value, error) {
	val, diags := e.expression.Value(nil)

	if diags.HasErrors() {

		return cty.NilVal, diags
	}

	return val, nil
}
