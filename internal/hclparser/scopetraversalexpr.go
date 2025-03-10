// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/zclconf/go-cty/cty"
)

// ScopeTraversalExpr resolves an HCL scope traversal (e.g. var.name or var.list[0]) against a variable store.
type ScopeTraversalExpr struct {
	variables  *HCLVars
	expression *hclsyntax.ScopeTraversalExpr
}

// NewScopeTraversalExpr creates a ScopeTraversalExpr for the given traversal expression and variable store.
func NewScopeTraversalExpr(expression *hclsyntax.ScopeTraversalExpr, hv *HCLVars) *ScopeTraversalExpr {
	return &ScopeTraversalExpr{
		variables:  hv,
		expression: expression,
	}
}

// VariableRef holds the resolved attribute name and optional index from a traversal expression.
type VariableRef struct {
	Attr  string
	Index *int64
}

// Parse walks the traversal segments, extracts the attribute name and optional index, and looks up the value.
func (ste *ScopeTraversalExpr) Parse() (cty.Value, error) {
	exprs, diags := hcl.AbsTraversalForExpr(ste.expression)
	if diags.HasErrors() {
		return cty.NilVal, cinzelerror.ProcessHCLDiags(diags)
	}

	variableRef := VariableRef{}

	for _, exp := range exprs {
		switch expressionType := exp.(type) {
		case hcl.TraverseRoot:
			// root segment (e.g. "var") is not used for variable lookup
		case hcl.TraverseAttr:
			variableRef.Attr = expressionType.Name
		case hcl.TraverseIndex:
			idx, _ := expressionType.Key.AsBigFloat().Int64()
			variableRef.Index = &idx
		default:
			return cty.NilVal, fmt.Errorf("unsupported %s", expressionType)
		}

	}

	variableValue, err := ste.variables.GetValue(variableRef.Attr, variableRef.Index)
	if err != nil {
		return cty.NilVal, err
	}

	return variableValue, nil
}
