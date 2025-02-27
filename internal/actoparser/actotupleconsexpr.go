// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"errors"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type ActoTupleConsExpr struct {
	expression *hclsyntax.TupleConsExpr
}

func NewActoTupleConsExpr(expression *hclsyntax.TupleConsExpr) *ActoTupleConsExpr {
	return &ActoTupleConsExpr{
		expression: expression,
	}
}

func (acto *ActoTupleConsExpr) Parse() (any, error) {
	list := struct {
		Grid              []int64
		AsString          []string
		AsUint64          []uint64
		AsInt64           []int64
		AsFloat64         []float64
		AsBool            []bool
		AsActoVariableRef []ActoVariableRef
		AsMap             []map[string]any
	}{
		Grid: []int64{0, 0, 0, 0, 0, 0, 0},
	}

	for _, expr := range acto.expression.Exprs {
		childActo := NewActo(expr)

		if err := childActo.Parse(); err != nil {
			return nil, err
		}

		switch childType := childActo.Result.(type) {
		case string:
			list.Grid[0] = 1
			list.AsString = append(list.AsString, childType)
		case uint64:
			list.Grid[1] = 1
			list.AsUint64 = append(list.AsUint64, childType)
		case int64:
			list.Grid[2] = 1
			list.AsInt64 = append(list.AsInt64, childType)
		case float64:
			list.Grid[3] = 1
			list.AsFloat64 = append(list.AsFloat64, childType)
		case bool:
			list.Grid[4] = 1
			list.AsBool = append(list.AsBool, childType)
		case ActoVariableRef:
			list.Grid[5] = 1
			list.AsActoVariableRef = append(list.AsActoVariableRef, childType)
		case map[string]cty.Value:
			list.Grid[6] = 1

			if list.AsMap == nil {
				list.AsMap = []map[string]any{}
			}

			setmap := make(map[string]any)
			for childKey, childVal := range childType {
				valueType := childVal.Type().FriendlyName()

				switch valueType {

				case cty.String.FriendlyName():
					setmap[childKey] = childVal.AsString()
				case cty.Number.FriendlyName():
					if !childVal.AsBigFloat().IsInt() {
						valueFloat, _ := childVal.AsBigFloat().Float64()
						setmap[childKey] = valueFloat
					} else {
						if childVal.AsBigFloat().Sign() == -1 {
							valueInt, _ := childVal.AsBigFloat().Int64()
							setmap[childKey] = valueInt
						} else {
							valueUint, _ := childVal.AsBigFloat().Uint64()
							setmap[childKey] = valueUint
						}
					}
				case cty.Bool.FriendlyName():
					setmap[childKey] = childVal.True()
				case cty.EmptyTuple.FriendlyName():
					var list []any
					valueIterator := childVal.ElementIterator()
					for valueIterator.Next() {
						_, valueElement := valueIterator.Element()
						val, err := CtyValueParser(valueElement)
						if err != nil {
							return nil, err
						}
						list = append(list, val)
					}
					setmap[childKey] = list
				}

			}
			list.AsMap = append(list.AsMap, setmap)
		default:
			return nil, errors.New("unkown ActoTupleConsExpr")
		}
	}

	var sum int64
	for _, grid := range list.Grid {
		sum += grid
	}

	if sum > 1 {
		return nil, errors.New("only one type of values is allowed in a list")
	}

	for idx, grid := range list.Grid {
		if grid != 1 {
			continue
		}

		switch idx {
		case 0:
			return list.AsString, nil
		case 1:
			return list.AsUint64, nil
		case 2:
			return list.AsInt64, nil
		case 3:
			return list.AsFloat64, nil
		case 4:
			return list.AsBool, nil
		case 5:
			return list.AsActoVariableRef, nil
		case 6:
			return list.AsMap, nil
		}
	}

	return nil, errors.New("unkown ActoTupleConsExpr")
}
