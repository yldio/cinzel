// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type WithConfig struct {
	Name  hcl.Expression `hcl:"name,attr"`
	Value hcl.Expression `hcl:"value,attr"`
}

type WithsConfig []*WithConfig

func (config *WithConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'name' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'name' must be a string")
	}
}

func (config *WithConfig) parseName() (*string, error) {
	acto := actoparser.NewActo(config.Name)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapName(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *WithConfig) unwrapValue(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'value' must be a string, number or boolean")
	case string:
		return &resultValue, nil
	case int64:
		return &resultValue, nil
	case uint64:
		return &resultValue, nil
	case float64:
		return &resultValue, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapValue(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'value' must be a string, number or boolean")
	}
}

func (config *WithConfig) parseValue() (any, error) {
	acto := actoparser.NewActo(config.Value)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapValue(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *WithsConfig) Parse() (*map[string]any, error) {
	if config == nil {
		return nil, nil
	}

	withs := make(map[string]any)

	for _, with := range *config {
		name, err := with.parseName()
		if err != nil {
			return nil, err
		}

		value, err := with.parseValue()
		if err != nil {
			return nil, err
		}

		withs[*name] = value
	}

	if len(withs) == 0 {
		return nil, nil
	}

	return &withs, nil
}
