// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/zclconf/go-cty/cty"
)

type ActoUnaryOpExpr struct {
	expression *hclsyntax.UnaryOpExpr
}

func NewActoUnaryOpExpr(expression *hclsyntax.UnaryOpExpr) *ActoUnaryOpExpr {
	return &ActoUnaryOpExpr{
		expression: expression,
	}
}

func (acto *ActoUnaryOpExpr) Parse() (any, error) {
	value, diags := acto.expression.Value(nil)
	if diags.HasErrors() {
		return nil, actoerrors.ProcessHCLDiags(diags)
	}

	valueType := value.Type().FriendlyName()

	switch valueType {
	case cty.String.FriendlyName():
		return value.AsString(), nil
	case cty.Number.FriendlyName():
		if !value.AsBigFloat().IsInt() {
			valueFloat, _ := value.AsBigFloat().Float64()
			return valueFloat, nil
		} else {
			if value.AsBigFloat().Sign() == -1 {
				valueInt, _ := value.AsBigFloat().Int64()
				return valueInt, nil
			} else {
				valueUint, _ := value.AsBigFloat().Uint64()
				return valueUint, nil
			}
		}
	case cty.Bool.FriendlyName():
		return value.True(), nil
	case cty.EmptyTuple.FriendlyName():
		var list []any
		valueIterator := value.ElementIterator()
		for valueIterator.Next() {
			_, valueElement := valueIterator.Element()
			val, err := CtyValueParser(valueElement)
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		}
		return list, nil
	default:
		return nil, fmt.Errorf("unkown type found %s", valueType)
	}

	return nil, nil
}
