// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

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

// Parse evaluates the template parts and returns the first resolved value.
func (te *TemplateExpr) Parse() (cty.Value, error) {
	for _, part := range te.expression.Parts {
		value, diag := part.Value(nil)
		if diag.HasErrors() {
			return cty.NilVal, errors.New(diag.Error())
		}

		switch value.Type() {
		case cty.String:
			return value, nil
		case cty.Number:
			return value, nil
		case cty.Bool:
			return value, nil
		default:
			return cty.NilVal, fmt.Errorf("unknown type found %s", value.Type().FriendlyName())
		}
	}

	return cty.NilVal, nil
}
