// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type Environment struct {
	Name string `hcl:"name,omitempty"`
	Url  string `hcl:"name,omitempty"`
}

type EnvironmentConfig struct {
	Name hcl.Expression `hcl:"name,attr"`
	Url  hcl.Expression `hcl:"url,attr"`
}

func (config *EnvironmentConfig) unwrapUrl(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapUrl(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'url' must be a string")
	}
}

func (config *EnvironmentConfig) ParseUrl() (*string, error) {
	acto := actoparser.NewActo(config.Url)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapUrl(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *EnvironmentConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
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

func (config *EnvironmentConfig) ParseName() (*string, error) {
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

func (config *EnvironmentConfig) Parse() (any, error) {
	if config == nil {
		return nil, nil
	}

	name, err := config.ParseName()
	if err != nil {
		return nil, err
	}

	url, err := config.ParseUrl()
	if err != nil {
		return nil, err
	}

	if name != nil && url != nil {
		environment := Environment{
			Name: *name,
			Url:  *url,
		}

		return environment, nil
	}

	return name, nil
}
