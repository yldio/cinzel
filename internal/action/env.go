// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type Env struct {
	Name  string `yaml:"-"`
	Value any    `yaml:"-"`
}

type Envs map[string]any

type EnvConfig struct {
	Name  hcl.Expression `hcl:"name,attr"`
	Value hcl.Expression `hcl:"value,attr"`
}

type EnvsConfig []*EnvConfig

func (config *EnvConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'name' must be value a number, string or boolean")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'name' must be value a number, string or boolean")
	}
}

func (config *EnvConfig) parseName() (*string, error) {
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

func (config *EnvConfig) unwrapValue(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'value' must be value a number, string or boolean")
	case string:
		return &resultValue, nil
	case uint64:
		return &resultValue, nil
	case int64:
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
		return nil, errors.New("attribute 'value' must be value a number, string or boolean")
	}
}

func (config *EnvConfig) parseValue() (any, error) {
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

func (config *EnvsConfig) Parse() (*Envs, error) {
	envs := make(Envs)

	for _, env := range *config {
		name, err := env.parseName()
		if err != nil {
			return nil, err
		}

		value, err := env.parseValue()
		if err != nil {
			return nil, err
		}

		envs[*name] = value
	}

	if len(envs) == 0 {
		return nil, nil
	}

	return &envs, nil
}
