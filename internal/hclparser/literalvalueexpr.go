// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// LiteralValueExpr wraps an HCL literal value (string, number, bool).
type LiteralValueExpr struct {
	expression *hclsyntax.LiteralValueExpr
}

// NewLiteralValueExpr creates a LiteralValueExpr for the given HCL literal.
func NewLiteralValueExpr(expression *hclsyntax.LiteralValueExpr) *LiteralValueExpr {
	return &LiteralValueExpr{
		expression: expression,
	}
}

// Parse returns the literal's cty value directly.
func (lve *LiteralValueExpr) Parse() (cty.Value, error) {
	return lve.expression.Val, nil
}
