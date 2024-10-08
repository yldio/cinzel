// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type Acto struct {
	expression hcl.Expression
	Result     any
}

func NewActo(expression hcl.Expression) *Acto {
	return &Acto{
		expression: expression,
	}
}

func NewActoFromResult(result any) *Acto {
	return &Acto{
		Result: result,
	}
}

func (acto *Acto) Parse() error {
	var expression cty.Value

	if acto.expression != nil {
		exp, diags := acto.expression.Value(nil)
		if diags.HasErrors() {
			// return diags
		}

		expression = exp
	}

	if expression == (cty.Value{}) || expression.IsNull() {
		return nil
	}

	switch expType := acto.expression.(type) {
	case *hclsyntax.LiteralValueExpr:
		value, err := NewActoLiteralValueExpr(expType).Parse()
		if err != nil {
			return err
		}

		acto.Result = value

		return nil
	case *hclsyntax.UnaryOpExpr:
		value, err := NewActoUnaryOpExpr(expType).Parse()
		if err != nil {
			return err
		}

		acto.Result = value

		return nil
	case *hclsyntax.ScopeTraversalExpr:
		value, err := NewActoScopeTraversalExpr(expType).Parse()
		if err != nil {
			return err
		}

		acto.Result = value

		return nil
	case *hclsyntax.TemplateExpr:
		value, err := NewActoTemplateExpr(expType).Parse()
		if err != nil {
			return err
		}

		acto.Result = value

		return nil
	case *hclsyntax.TupleConsExpr:
		value, err := NewActoTupleConsExpr(expType).Parse()
		if err != nil {
			return err
		}

		acto.Result = value

		return nil
	case hcl.Expression:
		value, err := NewActoExpression(expType).Parse()
		if err != nil {
			return err
		}

		acto.Result = value

		return nil
	default:
		return fmt.Errorf("missing hcl type found, found %s", expType)
	}
}

func CtyValueParser(value cty.Value) (any, error) {
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
}
