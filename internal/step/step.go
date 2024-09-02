// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package step

import (
	"github.com/yldio/acto/internal/action"
)

type Steps []Step

type Step struct {
	Id               string          `yaml:"id,omitempty"`
	If               *string         `yaml:"if,omitempty"`
	Name             *string         `yaml:"name,omitempty"`
	Uses             *string         `yaml:"uses,omitempty"`
	Run              *string         `yaml:"run,omitempty"`
	WorkingDirectory *string         `yaml:"working-directory,omitempty"`
	Shell            *string         `yaml:"shell,omitempty"`
	With             *map[string]any `yaml:"with,omitempty"`
	Env              *action.Env     `yaml:"env,omitempty"`
	ContinueOnError  *bool           `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes   *uint16         `yaml:"timeout-minutes,omitempty"`
}

type StepsConfig []StepConfig

type StepConfig struct {
	Id               string                         `hcl:"id,label"`
	If               *action.IfConfig               `hcl:"if,attr"`
	Name             *action.NameConfig             `hcl:"name,attr"`
	Uses             *action.UsesConfig             `hcl:"uses,block"`
	Run              *action.RunConfig              `hcl:"run,attr"`
	WorkingDirectory *action.WorkingDirectoryConfig `hcl:"working_directory,attr"`
	Shell            *action.ShellConfig            `hcl:"shell,attr"`
	With             action.WithsConfig             `hcl:"with,block"` // action.WithsConfig -> []*WithConfig
	Env              *action.EnvConfig              `hcl:"env,block"`
	ContinueOnError  *action.ContinueOnErrorConfig  `hcl:"continue_on_error,attr"`
	TimeoutMinutes   *action.TimeoutMinutesConfig   `hcl:"timeout_minutes,attr"`
}

func (config *StepConfig) parseId(step *Step) error {
	if config.Id == "" {
		return nil
	}

	step.Id = config.Id

	return nil
}

func (config *StepConfig) parseIf(step *Step) error {
	iF, err := config.If.Parse()
	if err != nil {
		return err
	}

	if iF == "" {
		return nil
	}

	step.If = &iF

	return nil
}

func (config *StepConfig) parseName(step *Step) error {
	name, err := config.Name.Parse()
	if err != nil {
		return err
	}

	if name == "" {
		return nil
	}

	step.Name = &name

	return nil
}

func (config *StepConfig) parseUses(step *Step) error {
	uses, err := config.Uses.Parse()
	if err != nil {
		return err
	}

	if uses == "" {
		return nil
	}

	step.Uses = &uses

	return nil
}

func (config *StepConfig) parseRun(step *Step) error {
	run, err := config.Run.Parse()
	if err != nil {
		return err
	}

	if run == "" {
		return nil
	}

	step.Run = &run

	return nil
}

func (config *StepConfig) parseWorkingDirectory(step *Step) error {
	workingDirectory, err := config.WorkingDirectory.Parse()
	if err != nil {
		return err
	}

	if workingDirectory == "" {
		return nil
	}

	step.WorkingDirectory = &workingDirectory

	return nil
}

func (config *StepConfig) parseShell(step *Step) error {
	shell, err := config.Shell.Parse()
	if err != nil {
		return err
	}

	if shell == "" {
		return nil
	}

	step.Shell = &shell

	return nil
}

func (config *StepConfig) parseWith(step *Step) error {
	with, err := config.With.Parse()
	if err != nil {
		return err
	}

	if with == nil {
		return nil
	}

	step.With = &with

	return nil
}

func (config *StepConfig) parseEnv(step *Step) error {
	env, err := config.Env.Parse()
	if err != nil {
		return err
	}

	if env == nil {
		return nil
	}

	step.Env = &env

	return nil
}

func (config *StepConfig) parseContinueOnError(step *Step) error {
	continueOnError, err := config.ContinueOnError.Parse()
	if err != nil {
		return err
	}

	if continueOnError == nil {
		return nil
	}

	step.ContinueOnError = continueOnError

	return nil
}

func (config *StepConfig) parseTimeoutMinutes(step *Step) error {
	timeoutMinutes, err := config.TimeoutMinutes.Parse()
	if err != nil {
		return err
	}

	if timeoutMinutes == nil {
		return nil
	}

	step.TimeoutMinutes = timeoutMinutes

	return nil
}

func (config *StepConfig) Parse() (Step, error) {
	if config == nil {
		return Step{}, nil
	}

	step := Step{}

	if err := config.parseId(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseIf(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseName(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseUses(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseRun(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseWorkingDirectory(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseShell(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseWith(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseEnv(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseContinueOnError(&step); err != nil {
		return Step{}, err
	}

	if err := config.parseTimeoutMinutes(&step); err != nil {
		return Step{}, err
	}

	return step, nil
}

func (config *StepsConfig) Parse() (Steps, error) {
	steps := Steps{}

	for _, step := range *config {
		parsedStep, err := step.Parse()
		if err != nil {
			return Steps{}, err
		}

		steps = append(steps, parsedStep)
	}

	return steps, nil
}
