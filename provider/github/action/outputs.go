// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type Outputs map[string]string

type OutputConfig struct {
	Name  string         `hcl:"_,label"`
	Value hcl.Expression `hcl:"value,attr"`
}

type OutputsConfig []*OutputConfig

func (config *OutputConfig) unwrapRunners(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'runners' must be a string, number, bool")
	case string:
		return &resultValue, nil
	case int64:
		val := fmt.Sprintf("%v", resultValue)
		return &val, nil
	case uint64:
		val := fmt.Sprintf("%v", resultValue)
		return &val, nil
	case float64:
		val := fmt.Sprintf("%v", resultValue)
		return &val, nil
	case bool:
		val := fmt.Sprintf("%v", resultValue)
		return &val, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapRunners(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'runners' must be a string, number, bool")
	}
}

func (config *OutputConfig) Parse() (*string, error) {
	if config == nil {
		return nil, nil
	}

	if config.Name == "" {
		return nil, errors.New("missing 'output' identifier")
	}

	acto := actoparser.NewActo(config.Value)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapRunners(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *OutputsConfig) Parse() (*Outputs, error) {
	if config == nil {
		return nil, nil
	}

	outputs := Outputs{}

	for _, output := range *config {
		val, err := output.Parse()
		if err != nil {
			return nil, err
		}

		outputs[output.Name] = *val
	}

	return &outputs, nil
}
