// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

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

func (envs *Envs) Decode(body *hclwrite.Body, attr string) error {

	for key, value := range *envs {
		if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
			body.AppendNewline()
		}

		envBlock := body.AppendNewBlock(attr, nil)
		envBody := envBlock.Body()

		envBody.SetAttributeValue("name", cty.StringVal(key))

		switch v := value.(type) {
		case string:
			envBody.SetAttributeValue("value", cty.StringVal(v))
		case bool:
			envBody.SetAttributeValue("value", cty.BoolVal(v))
		case uint64:
			envBody.SetAttributeValue("value", cty.NumberUIntVal(v))
		case int64:
			envBody.SetAttributeValue("value", cty.NumberIntVal(v))
		case float64:
			envBody.SetAttributeValue("value", cty.NumberFloatVal(v))
		default:
			return errors.New("unkown dealt type")
		}
	}

	return nil
}
