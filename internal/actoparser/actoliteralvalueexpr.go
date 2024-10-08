// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type ActoLiteralValueExpr struct {
	expression *hclsyntax.LiteralValueExpr
}

func NewActoLiteralValueExpr(expression *hclsyntax.LiteralValueExpr) *ActoLiteralValueExpr {
	return &ActoLiteralValueExpr{
		expression: expression,
	}
}

func (acto *ActoLiteralValueExpr) Parse() (any, error) {
	return CtyValueParser(acto.expression.Val)
}
