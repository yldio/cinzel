// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// TemplateExpr evaluates an HCL template string expression.
type TemplateExpr struct {
	expression *hclsyntax.TemplateExpr
}

// NewTemplateExpr creates a TemplateExpr for the given HCL template.
func NewTemplateExpr(expression *hclsyntax.TemplateExpr) *TemplateExpr {
	return &TemplateExpr{
		expression: expression,
	}
}

// Parse evaluates the template and returns its value. A template of one part
// keeps that part's type, so a bare number or bool is not turned into a string.
//
// Returning on the first part truncated every template that has more than one:
// "prefix ${1} suffix" came out as "prefix ", written to the file with no
// error and no warning.
func (te *TemplateExpr) Parse() (cty.Value, error) {
	if len(te.expression.Parts) == 0 {
		return cty.NilVal, nil
	}

	if len(te.expression.Parts) == 1 {
		value, diag := te.expression.Parts[0].Value(nil)

		if diag.HasErrors() {
			return cty.NilVal, errors.New(diag.Error())
		}

		switch value.Type() {
		case cty.String, cty.Number, cty.Bool:
			return value, nil
		default:
			return cty.NilVal, fmt.Errorf("unknown type found %s", value.Type().FriendlyName())
		}
	}

	// Several parts always make a string, and joining them is what the
	// template expression itself does, down to how a number is spelled.
	value, diag := te.expression.Value(nil)

	if diag.HasErrors() {
		return cty.NilVal, errors.New(diag.Error())
	}

	return value, nil
}
