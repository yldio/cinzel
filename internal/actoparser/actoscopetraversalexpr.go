// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/acto/internal/actoerrors"
)

type ActoScopeTraversalExpr struct {
	expression *hclsyntax.ScopeTraversalExpr
}

func NewActoScopeTraversalExpr(expression *hclsyntax.ScopeTraversalExpr) *ActoScopeTraversalExpr {
	return &ActoScopeTraversalExpr{
		expression: expression,
	}
}

type ActoVariableRef struct {
	Name  string
	Attr  string
	Index *int64
}

func (acto *ActoScopeTraversalExpr) Parse() (any, error) {
	exprs, diags := hcl.AbsTraversalForExpr(acto.expression)
	if diags.HasErrors() {
		return nil, actoerrors.ProcessHCLDiags(diags)
	}

	actoVariableRef := ActoVariableRef{}

	for _, exp := range exprs {
		switch expressionType := exp.(type) {
		case hcl.TraverseRoot:
			actoVariableRef.Name = expressionType.Name
		case hcl.TraverseAttr:
			actoVariableRef.Attr = expressionType.Name
		case hcl.TraverseIndex:
			idx, _ := expressionType.Key.AsBigFloat().Int64()
			actoVariableRef.Index = &idx
		default:
			return nil, fmt.Errorf("unsupported %s", expressionType)
		}

	}

	return actoVariableRef, nil
}
