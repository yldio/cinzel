// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package hclparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// BinaryOpExpr evaluates an HCL binary operation (e.g. +, -, ==) with variable support.
type BinaryOpExpr struct {
	variables  *HCLVars
	expression *hclsyntax.BinaryOpExpr
}

// NewBinaryOpExpr creates a BinaryOpExpr for the given expression and variable store.
func NewBinaryOpExpr(expression *hclsyntax.BinaryOpExpr, hv *HCLVars) *BinaryOpExpr {

	return &BinaryOpExpr{
		variables:  hv,
		expression: expression,
	}
}

// Parse evaluates the binary operation and returns the resulting cty value.
func (boe *BinaryOpExpr) Parse() (cty.Value, error) {
	lhs, err := boe.parseLHS()
	if err != nil {

		return cty.NilVal, err
	}

	rhs, err := boe.parseRHS()
	if err != nil {

		return cty.NilVal, err
	}

	switch boe.expression.Op {
	case hclsyntax.OpAdd:
		lVal, _ := lhs.AsBigFloat().Int64()
		rVal, _ := rhs.AsBigFloat().Int64()

		return cty.NumberIntVal(lVal + rVal), nil
	case hclsyntax.OpSubtract:
		lVal, _ := lhs.AsBigFloat().Int64()
		rVal, _ := rhs.AsBigFloat().Int64()

		return cty.NumberIntVal(lVal - rVal), nil
	case hclsyntax.OpMultiply:
		lVal, _ := lhs.AsBigFloat().Int64()
		rVal, _ := rhs.AsBigFloat().Int64()

		return cty.NumberIntVal(lVal * rVal), nil
	case hclsyntax.OpDivide:
		if rhs.IsNull() || rhs.AsBigFloat().Sign() == 0 {

			return cty.NilVal, fmt.Errorf("division by zero")
		}

		lVal, _ := lhs.AsBigFloat().Float64()
		rVal, _ := rhs.AsBigFloat().Float64()

		return cty.NumberFloatVal(lVal / rVal), nil
	case hclsyntax.OpEqual:
		return cty.BoolVal(lhs.RawEquals(rhs)), nil
	case hclsyntax.OpNotEqual:
		return cty.BoolVal(!lhs.RawEquals(rhs)), nil
	case hclsyntax.OpGreaterThan:
		return cty.BoolVal(lhs.AsBigFloat().Cmp(rhs.AsBigFloat()) > 0), nil
	case hclsyntax.OpGreaterThanOrEqual:
		return cty.BoolVal(lhs.AsBigFloat().Cmp(rhs.AsBigFloat()) >= 0), nil
	case hclsyntax.OpLessThan:
		return cty.BoolVal(lhs.AsBigFloat().Cmp(rhs.AsBigFloat()) < 0), nil
	case hclsyntax.OpLessThanOrEqual:
		return cty.BoolVal(lhs.AsBigFloat().Cmp(rhs.AsBigFloat()) <= 0), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported binary operator")
	}
}

func (boe *BinaryOpExpr) parseLHS() (cty.Value, error) {

	return parseExpression(boe.expression.LHS, boe.variables)
}

func (boe *BinaryOpExpr) parseRHS() (cty.Value, error) {

	return parseExpression(boe.expression.RHS, boe.variables)
}

func parseExpression(expr hclsyntax.Expression, hv *HCLVars) (cty.Value, error) {
	switch e := expr.(type) {
	case *hclsyntax.LiteralValueExpr:
		return e.Val, nil
	case *hclsyntax.UnaryOpExpr:
		val, diags := e.Value(nil)

		if diags.HasErrors() {

			return cty.NilVal, diags
		}

		return val, nil
	case *hclsyntax.ScopeTraversalExpr:
		return NewScopeTraversalExpr(e, hv).Parse()
	case *hclsyntax.ParenthesesExpr:
		return NewExpression(e.Expression).Parse()
	case *hclsyntax.BinaryOpExpr:
		return NewBinaryOpExpr(e, hv).Parse()
	default:
		return cty.NilVal, fmt.Errorf("unsupported expression type: %T", expr)
	}
}
