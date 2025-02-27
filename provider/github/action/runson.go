// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type RunsOn struct {
	Group  any `yaml:"group,omitempty"`
	Labels any `yaml:"labels,omitempty"`
}

type RunsOnConfig struct {
	Runners hcl.Expression `hcl:"runners,attr"`
	Group   hcl.Expression `hcl:"group,attr"`
	Labels  hcl.Expression `hcl:"labels,attr"`
}

func (config *RunsOnConfig) unwrapLabels(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case []string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapLabels(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'labels' must be a string or a list of strings")
	}
}

func (config *RunsOnConfig) ParseLabels() (any, error) {
	acto := actoparser.NewActo(config.Labels)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapLabels(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *RunsOnConfig) unwrapGroup(acto *actoparser.Acto) (*string, error) {
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
		return nil, errors.New("attribute 'group' must be a string")
	}
}

func (config *RunsOnConfig) ParseGroup() (*string, error) {
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

func (config *RunsOnConfig) unwrapRunners(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case []string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapRunners(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'runners' must be a string or list of strings")
	}
}

func (config *RunsOnConfig) ParseRunners() (any, error) {
	acto := actoparser.NewActo(config.Runners)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapRunners(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *RunsOnConfig) Parse() (any, error) {
	if config == nil {
		return nil, nil
	}

	runners, err := config.ParseRunners()
	if err != nil {
		return nil, err
	}

	group, err := config.ParseGroup()
	if err != nil {
		return nil, err
	}

	labels, err := config.ParseLabels()
	if err != nil {
		return nil, err
	}

	if runners != nil && (group != nil || labels != nil) {
		return nil, errors.New("can only have 'runners' or a combination of 'groups' with optional 'labels'")
	}

	if runners != nil {
		return runners, nil
	}

	if group == nil {
		return nil, errors.New("if 'runners' is not set, requires a combination of 'groups' with optional 'labels'")
	}

	runsOn := RunsOn{
		Group: group,
	}

	if labels != nil {
		runsOn.Labels = labels
	}

	return runsOn, nil
}
