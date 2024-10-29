// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
)

type Run struct {
	Shell            *string `yaml:"shell,omitempty" hcl:"shell"`
	WorkingDirectory *string `yaml:"working-directory,omitempty" hcl:"working_directory"`
}

type Defaults struct {
	Run *Run `yaml:"run,omitempty" hcl:"run"`
}

type DefaultsRunConfig struct {
	Shell            hcl.Expression `hcl:"shell,attr"`
	WorkingDirectory hcl.Expression `hcl:"working_directory,attr"`
}

type DefaultsConfig struct {
	Run *DefaultsRunConfig `hcl:"run,block"`
}

func (config *DefaultsRunConfig) unwrapWorkingDirectory(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapWorkingDirectory(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'working_directory' must be value a string")
	}
}

func (config *DefaultsRunConfig) parseWorkingDirectory() (*string, error) {
	acto := actoparser.NewActo(config.WorkingDirectory)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapWorkingDirectory(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *DefaultsRunConfig) unwrapShell(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapShell(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'shell' must be value a string")
	}
}

func (config *DefaultsRunConfig) parseShell() (*string, error) {
	acto := actoparser.NewActo(config.Shell)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapShell(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *DefaultsConfig) Parse() (*Defaults, error) {
	if config == nil {
		return nil, nil
	}

	if config.Run == nil {
		return nil, errors.New("block 'defaults' required a 'run' block")
	}

	defaults := Defaults{
		Run: &Run{},
	}

	shell, err := config.Run.parseShell()
	if err != nil {
		return nil, err
	}

	if shell != nil {
		defaults.Run.Shell = shell
	}

	workingDirectory, err := config.Run.parseWorkingDirectory()
	if err != nil {
		return nil, err
	}

	if workingDirectory != nil {
		defaults.Run.WorkingDirectory = workingDirectory
	}

	if shell == nil && workingDirectory == nil {
		return nil, errors.New("block 'defaults' required a 'run' block to have 'shell' or/and 'working_directory'")
	}

	return &defaults, nil
}
