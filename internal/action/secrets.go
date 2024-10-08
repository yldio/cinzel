// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type SecretConfig struct {
	Name  hcl.Expression `hcl:"name,attr"`
	Value hcl.Expression `hcl:"value,attr"`
}

type SecretsConfig []*SecretConfig

func (config *SecretConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
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

func (config *SecretConfig) parseName() (*string, error) {
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

func (config *SecretConfig) unwrapValue(acto *actoparser.Acto) (any, error) {
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

func (config *SecretConfig) parseValue() (any, error) {
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

func (config *SecretsConfig) Parse() (*map[string]any, error) {
	if config == nil {
		return nil, nil
	}

	secrets := make(map[string]any)

	for _, secret := range *config {
		name, err := secret.parseName()
		if err != nil {
			return nil, err
		}

		value, err := secret.parseValue()
		if err != nil {
			return nil, err
		}

		secrets[*name] = value
	}

	if len(secrets) == 0 {
		return nil, nil
	}

	return &secrets, nil
}
