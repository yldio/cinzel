// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type UsesConfig struct {
	Action  hcl.Expression `hcl:"action,attr"`
	Version hcl.Expression `hcl:"version,attr"`
}

func (config *UsesConfig) unwrapVersion(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'action' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapVersion(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'version' must be a string")
	}
}

func (config *UsesConfig) parseVersion() (*string, error) {
	acto := actoparser.NewActo(config.Version)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapVersion(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *UsesConfig) unwrapAction(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'action' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapAction(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'action' must be a string")
	}
}

func (config *UsesConfig) parseAction() (*string, error) {
	acto := actoparser.NewActo(config.Action)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapAction(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *UsesConfig) Parse() (*string, error) {
	if config == nil {
		return nil, nil
	}

	action, err := config.parseAction()
	if err != nil {
	}

	version, err := config.parseVersion()
	if err != nil {
	}

	uses := fmt.Sprintf("%s@%s", *action, *version)

	return &uses, nil
}
