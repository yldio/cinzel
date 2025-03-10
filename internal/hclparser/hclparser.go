// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// HCLParser evaluates a single HCL expression, resolving variable references.
type HCLParser struct {
	variables  *HCLVars
	expression hcl.Expression
	result     cty.Value
}

// Variables returns the variable store used by this parser.
func (hp *HCLParser) Variables() *HCLVars {
	return hp.variables
}

// Result returns the evaluated cty value after Parse has been called.
func (hp *HCLParser) Result() cty.Value {
	return hp.result
}

// New creates an HCLParser for the given expression and variable store.
func New(expression hcl.Expression, hv *HCLVars) *HCLParser {
	return &HCLParser{
		variables:  hv,
		expression: expression,
	}
}

// Parse evaluates the expression and stores the result.
func (hp *HCLParser) Parse() error {
	var expression cty.Value

	if hp.expression != nil {
		expression, _ = hp.expression.Value(nil)
	}

	if expression.IsNull() {
		return nil
	}

	switch expType := hp.expression.(type) {
	case *hclsyntax.LiteralValueExpr:
		value, err := NewLiteralValueExpr(expType).Parse()
		if err != nil {
			return err
		}

		hp.result = value

		return nil
	case *hclsyntax.UnaryOpExpr:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		hp.result = value
		return nil
	case *hclsyntax.ScopeTraversalExpr:
		value, err := NewScopeTraversalExpr(expType, hp.variables).Parse()
		if err != nil {
			return err
		}

		hp.result = value

		return nil
	case *hclsyntax.TemplateExpr:
		value, err := NewTemplateExpr(expType).Parse()
		if err != nil {
			return err
		}

		hp.result = value

		return nil
	case *hclsyntax.TupleConsExpr:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		hp.result = value
		return nil
	case *hclsyntax.BinaryOpExpr:
		value, err := NewBinaryOpExpr(expType, hp.variables).Parse()
		if err != nil {
			return err
		}

		hp.result = value
		return nil
	case *hclsyntax.ConditionalExpr:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		hp.result = value
		return nil
	case *hclsyntax.ForExpr:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		hp.result = value
		return nil
	case *hclsyntax.RelativeTraversalExpr:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		hp.result = value
		return nil
	case *hclsyntax.ObjectConsExpr:
		value, err := NewExpression(expType).Parse()
		if err != nil {
			return err
		}

		hp.result = value

		return nil
	case *hclsyntax.FunctionCallExpr:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return diags
		}

		hp.result = value
		return nil
	default:
		return fmt.Errorf("missing hcl type found, found %s", expType)
	}
}
