// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type actoValue cty.Value

func (actoValue *actoValue) Parse(allowedTypes []string) (any, error) {
	value := cty.Value(*actoValue)

	valueType := value.Type().FriendlyName()

	if !slices.Contains(allowedTypes, valueType) {
		return nil, fmt.Errorf("allowed types only %s", strings.Join(allowedTypes, ","))
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
	default:
		return nil, errors.New("missing case found")
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
