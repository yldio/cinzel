// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type ctyVal cty.Value

// Parse converts the underlying cty value into a Go value constrained by allowed type names.
func (ctyVal *ctyVal) Parse(allowedTypes []string) (any, error) {
	value := cty.Value(*ctyVal)

	valueType := value.Type().FriendlyName()

	allowedTypes = append(allowedTypes, "dynamic")

	if len(allowedTypes) > 1 && !slices.Contains(allowedTypes, valueType) {
		return nil, fmt.Errorf("%s only allows types %s", value, strings.Join(allowedTypes, ","))
	}

	switch valueType {
	case cty.String.FriendlyName():
		return ctyVal.ParseAsString()
	case cty.Number.FriendlyName():
		return ctyVal.ParseAsNumber()
	case cty.Bool.FriendlyName():
		return ctyVal.ParseAsBool()
	case cty.EmptyTuple.FriendlyName():
		return ctyVal.ParseAsTuple()
	case cty.DynamicPseudoType.FriendlyName():
		return nil, nil
	default:
		return nil, errors.New("missing cty type found")
	}
}

// ParseAsString converts the underlying cty value to a string.
func (ctyVal *ctyVal) ParseAsString() (any, error) {
	var val string
	value := cty.Value(*ctyVal)

	if err := gocty.FromCtyValue(value, &val); err != nil {
		return "", err
	}

	return val, nil
}

// ParseAsNumber converts the underlying cty value to an integer or float.
func (ctyVal *ctyVal) ParseAsNumber() (any, error) {
	var intVal int32
	value := cty.Value(*ctyVal)

	err := gocty.FromCtyValue(value, &intVal)
	if err != nil {
		var floatVal float32

		err := gocty.FromCtyValue(value, &floatVal)
		if err != nil {
			return nil, err
		}

		return floatVal, nil
	}

	return intVal, nil
}

// ParseAsBool converts the underlying cty value to a boolean.
func (ctyVal *ctyVal) ParseAsBool() (any, error) {
	var val bool
	value := cty.Value(*ctyVal)

	err := gocty.FromCtyValue(value, &val)
	if err != nil {
		return nil, err
	}

	return val, nil
}

// ParseAsTuple converts the underlying cty tuple value into a string slice.
func (ctyVal *ctyVal) ParseAsTuple() (any, error) {
	var val []string
	value := cty.Value(*ctyVal)

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

// ParseCtyValue converts a cty.Value to a native Go value, restricted to the given allowed type names.
func ParseCtyValue(value cty.Value, allowedTypes []string) (any, error) {
	ctyVal := ctyVal(value)

	return ctyVal.Parse(allowedTypes)
}
