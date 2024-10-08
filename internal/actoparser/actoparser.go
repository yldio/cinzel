// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package actoparser

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type ActoTraversal struct {
	Type  string
	Attr  string
	Index *int64
}

type actoValue cty.Value

func (actoValue *actoValue) Parse(allowedTypes []string) (any, error) {
	value := cty.Value(*actoValue)

	valueType := value.Type().FriendlyName()

	allowedTypes = append(allowedTypes, "dynamic")

	if len(allowedTypes) > 1 && !slices.Contains(allowedTypes, valueType) {
		return nil, fmt.Errorf("%s only allows types %s", value, strings.Join(allowedTypes, ","))
	}

	switch valueType {
	case cty.String.FriendlyName():
		return actoValue.ParseAsString()
	case cty.Number.FriendlyName():
		return actoValue.ParseAsNumber()
	case cty.Bool.FriendlyName():
		return actoValue.ParseAsBool()
	case cty.EmptyTuple.FriendlyName():
		return actoValue.ParseAsTuple()
	case cty.DynamicPseudoType.FriendlyName():
		return nil, nil
	default:
		return nil, errors.New("missing cty type found")
	}
}

func (actoValue *actoValue) ParseAsString() (any, error) {
	var val string
	value := cty.Value(*actoValue)

	if err := gocty.FromCtyValue(value, &val); err != nil {
		return "", err
	}

	return val, nil
}

func (actoValue *actoValue) ParseAsNumber() (any, error) {
	var val int32
	value := cty.Value(*actoValue)

	err := gocty.FromCtyValue(value, &val)
	if err != nil {
		var val float32

		err := gocty.FromCtyValue(value, &val)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

func (actoValue *actoValue) ParseAsBool() (any, error) {
	var val bool
	value := cty.Value(*actoValue)

	err := gocty.FromCtyValue(value, &val)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (actoValue *actoValue) ParseAsTuple() (any, error) {
	var val []string
	value := cty.Value(*actoValue)

	for _, item := range value.AsValueSlice() {
		var itemVal string

		err := gocty.FromCtyValue(item, &itemVal)
		if err != nil {
			return nil, err
		}

		val = append(val, itemVal)
	}

	return val, nil
}

func ParseCtyValue(value cty.Value, allowedTypes []string) (any, error) {
	actoValue := actoValue(value)

	return actoValue.Parse(allowedTypes)
}

type ParseHclExpressioner struct {
	Expression   hcl.Expression
	AllowedTypes []string
}

func (phe *ParseHclExpressioner) parseHclScopeTraversalExpr(expression *hclsyntax.ScopeTraversalExpr) (any, error) {
	exprs, diags := hcl.AbsTraversalForExpr(expression)
	if diags.HasErrors() {
		return nil, actoerrors.ProcessHCLDiags(diags)
	}

	actoTraversal := ActoTraversal{}

	for _, traverser := range exprs {
		switch traverserType := traverser.(type) {
		case hcl.TraverseRoot:
			actoTraversal.Type = traverserType.Name
		case hcl.TraverseAttr:
			actoTraversal.Attr = traverserType.Name
		case hcl.TraverseIndex:
			idx, _ := traverserType.Key.AsBigFloat().Int64()
			actoTraversal.Index = &idx
		default:
			return nil, errors.New("unsupported")
		}
	}

	if actoTraversal.Type != "variable" {
		return actoTraversal, nil
	}

	vars := variables.Instance()

	if actoTraversal.Index == nil {
		val, err := vars.GetValueByKey(actoTraversal.Attr)
		if err != nil {
			return nil, err
		}

		return val, nil
	} else {
		val, err := vars.GetValueByIndex(actoTraversal.Attr, *actoTraversal.Index)
		if err != nil {
			return nil, err
		}

		return val, nil
	}
}

func (phe *ParseHclExpressioner) parseHclLiteralValueExpr(expression *hclsyntax.LiteralValueExpr) (any, error) {
	return ParseCtyValue(expression.Val, phe.AllowedTypes)
}

func (phe *ParseHclExpressioner) parseHclTemplateExpr(expression *hclsyntax.TemplateExpr) (any, error) {
	var list []any
	for _, part := range expression.Parts {
		switch expType := part.(type) {
		case *hclsyntax.LiteralValueExpr:
			val, err := phe.parseHclLiteralValueExpr(expType)
			if err != nil {
				return nil, err
			}

			list = append(list, val)
		default:
			return nil, errors.New("missing templateExpr found")
		}
	}

	if len(list) == 0 {
		return nil, errors.New("templateExpr is empty")
	}

	return list, nil
}

func ParseHclExpression(parseHclExpressioner ParseHclExpressioner) (any, error) {
	if parseHclExpressioner.Expression == nil {
		return nil, nil
	}

	switch expType := parseHclExpressioner.Expression.(type) {
	case *hclsyntax.LiteralValueExpr:
		return parseHclExpressioner.parseHclLiteralValueExpr(expType)
	case *hclsyntax.ScopeTraversalExpr:
		return parseHclExpressioner.parseHclScopeTraversalExpr(expType)
	case *hclsyntax.TemplateExpr:
		return parseHclExpressioner.parseHclTemplateExpr(expType)
	case hcl.Expression:
		value, diags := expType.Value(nil)
		if diags.HasErrors() {
			return nil, actoerrors.ProcessHCLDiags(diags)
		}
		return ParseCtyValue(value, parseHclExpressioner.AllowedTypes)
	default:
		return nil, errors.New("missing hcl type found")
	}
}

func GetAllowedTypesFromTag(config any, fieldName string) ([]string, error) {
	if fieldName == "" {
		return nil, nil
	}

	t := reflect.TypeOf(config)

	field, _ := t.FieldByName(fieldName)
	tag := field.Tag.Get("acto")

	if tag == "" {
		return []string{}, nil
	}

	tags := strings.Split(tag, ",")
	var allowedTypes []string

	for _, tag := range tags {
		var allowedType string

		switch tag {
		case "string":
			allowedType = cty.String.FriendlyName()
		case "number":
			allowedType = cty.Number.FriendlyName()
		case "bool":
			allowedType = cty.Bool.FriendlyName()
		case "tuple":
			allowedType = cty.EmptyTuple.FriendlyName()
		default:
			return []string{}, errors.New("unknown tag type")
		}

		allowedTypes = append(allowedTypes, allowedType)
	}

	return allowedTypes, nil
}

func ParseHclValue(config any, expression hcl.Expression, valueName string) (any, error) {
	allowedTypes, err := GetAllowedTypesFromTag(config, valueName)
	if err != nil {
		return nil, err
	}

	parseHclExpressioner := ParseHclExpressioner{
		Expression:   expression,
		AllowedTypes: allowedTypes,
	}

	val, err := ParseHclExpression(parseHclExpressioner)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func ParseAsString(config any) (*string, error) {
	if config == nil {
		return nil, nil
	}

	var content string
	var ok bool

	switch value := config.(type) {
	case []any:
		content, ok = value[0].(string)
	case any:
		content, ok = value.(string)
	}

	if !ok {
		return nil, errors.New("could not parse as a string")
	}

	return &content, nil
}
