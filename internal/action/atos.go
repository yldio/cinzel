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

type AtosValue cty.Value

func (atosValue *AtosValue) Parse(allowedTypes []string) (any, error) {
	value := cty.Value(*atosValue)

	valueType := value.Type().FriendlyName()

	if !slices.Contains(allowedTypes, valueType) {
		return nil, fmt.Errorf("allowed types only %s", strings.Join(allowedTypes, ","))
	}

	switch valueType {
	case cty.String.FriendlyName():
		return atosValue.ParseAsString()
	case cty.Number.FriendlyName():
		return atosValue.ParseAsNumber()
	case cty.Bool.FriendlyName():
		return atosValue.ParseAsBool()
	case cty.EmptyTuple.FriendlyName():
		return atosValue.ParseAsTuple()
	default:
		return nil, errors.New("missing case found")
	}
}

func (atosValue *AtosValue) ParseAsString() (any, error) {
	var val string
	value := cty.Value(*atosValue)

	if err := gocty.FromCtyValue(value, &val); err != nil {
		return "", err
	}

	return val, nil
}

func (atosValue *AtosValue) ParseAsNumber() (any, error) {
	var val int32
	value := cty.Value(*atosValue)

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

func (atosValue *AtosValue) ParseAsBool() (any, error) {
	var val bool
	value := cty.Value(*atosValue)

	err := gocty.FromCtyValue(value, &val)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (atosValue *AtosValue) ParseAsTuple() (any, error) {
	var val []string
	value := cty.Value(*atosValue)

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
	atosValue := AtosValue(value)

	return atosValue.Parse(allowedTypes)
}
