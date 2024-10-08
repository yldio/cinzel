// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type Concurrency struct {
	Group            *string `yaml:"group,omitempty"`
	CancelInProgress *bool   `yaml:"cancel-in-progress,omitempty"`
}

type ConcurrencyConfig struct {
	Group            hcl.Expression `hcl:"group,attr"`
	CancelInProgress hcl.Expression `hcl:"cancel_in_progress,attr"`
}

func (config *ConcurrencyConfig) unwrapCancelInProgress(acto *actoparser.Acto) (*bool, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapCancelInProgress(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'cancel_in_progress' must be value a bool")
	}
}

func (config *ConcurrencyConfig) parseCancelInProgress() (*bool, error) {
	acto := actoparser.NewActo(config.CancelInProgress)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapCancelInProgress(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *ConcurrencyConfig) unwrapGroup(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapGroup(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'group' must be value a string")
	}
}

func (config *ConcurrencyConfig) parseGroup() (*string, error) {
	acto := actoparser.NewActo(config.Group)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapGroup(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *ConcurrencyConfig) Parse() (*Concurrency, error) {
	if config == nil {
		return nil, nil
	}

	concurrency := Concurrency{}

	group, err := config.parseGroup()
	if err != nil {
		return nil, err
	}

	if group != nil {
		concurrency.Group = group
	}

	cancelInProgress, err := config.parseCancelInProgress()
	if err != nil {
		return nil, err
	}

	if cancelInProgress != nil {
		concurrency.CancelInProgress = cancelInProgress
	}

	if group == nil && cancelInProgress == nil {
		return nil, errors.New("block 'concurrency' required a 'group' or/and 'cancel_in_progress' attributes")
	}

	if group == nil && cancelInProgress != nil {
		return nil, errors.New("block 'concurrency' required a 'group' or/and 'cancel_in_progress' attributes")
	}

	return &concurrency, nil
}
