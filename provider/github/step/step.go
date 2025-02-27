// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package step

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/yldio/acto/provider/github/action"
	"github.com/zclconf/go-cty/cty"
)

type Steps map[string]*Step

type Step struct {
	Identifier       string          `yaml:"-"`
	Id               *string         `yaml:"id,omitempty" hcl:"id"`
	If               *string         `yaml:"if,omitempty" hcl:"if"`
	Name             *string         `yaml:"name,omitempty" hcl:"name"`
	Uses             *string         `yaml:"uses,omitempty" hcl:"uses"`
	Run              *string         `yaml:"run,omitempty" hcl:"run"`
	WorkingDirectory *string         `yaml:"working-directory,omitempty" hcl:"working_directory"`
	Shell            *string         `yaml:"shell,omitempty" hcl:"shell"`
	With             *map[string]any `yaml:"with,omitempty" hcl:"with"`
	Env              *action.Envs    `yaml:"env,omitempty" hcl:"env"`
	ContinueOnError  any             `yaml:"continue-on-error,omitempty" hcl:"continue_on_error"`
	TimeoutMinutes   *uint64         `yaml:"timeout-minutes,omitempty" hcl:"timeout_minutes"`
}

type StepsConfig []StepConfig

type StepConfig struct {
	Identifier       string             `hcl:"id,label"`
	Id               hcl.Expression     `hcl:"id,attr"`
	IgnoreId         hcl.Expression     `hcl:"ignore_id,attr"`
	If               hcl.Expression     `hcl:"if,attr"`
	Name             hcl.Expression     `hcl:"name,attr"`
	Uses             *action.UsesConfig `hcl:"uses,block"`
	Run              hcl.Expression     `hcl:"run,attr"`
	WorkingDirectory hcl.Expression     `hcl:"working_directory,attr"`
	Shell            hcl.Expression     `hcl:"shell,attr"`
	With             action.WithsConfig `hcl:"with,block"`
	Env              action.EnvsConfig  `hcl:"env,block"`
	ContinueOnError  hcl.Expression     `hcl:"continue_on_error,attr"`
	TimeoutMinutes   hcl.Expression     `hcl:"timeout_minutes,attr"`
}

func (config *StepConfig) unwrapTimeoutMinutes(acto *actoparser.Acto) (*uint64, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case int64:
		if resultValue < 0 {
			return nil, errors.New("attribute 'timeout_minutes' must be a positive number")
		}

		val := uint64(resultValue)
		return &val, nil
	case uint64:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapTimeoutMinutes(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'timeout_minutes' must be a positive number")
	}
}

func (config *StepConfig) parseTimeoutMinutes() (*uint64, error) {
	acto := actoparser.NewActo(config.TimeoutMinutes)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapTimeoutMinutes(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *StepConfig) unwrapContinueOnError(acto *actoparser.Acto) (any, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case bool:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapContinueOnError(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'continue_on_error' must be a string or a boolean")
	}
}

func (config *StepConfig) parseContinueOnError() (any, error) {
	acto := actoparser.NewActo(config.ContinueOnError)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapContinueOnError(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *StepConfig) parseEnvs() (*action.Envs, error) {
	env, err := config.Env.Parse()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (config *StepConfig) parseWith() (*map[string]any, error) {
	value, err := config.With.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in with: %w", err)
	}

	return value, nil
}

func (config *StepConfig) unwrapShell(acto *actoparser.Acto) (*string, error) {
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
		return nil, errors.New("attribute 'shell' must be a string")
	}
}

func (config *StepConfig) parseShell() (*string, error) {
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

func (config *StepConfig) unwrapWorkingDirectory(acto *actoparser.Acto) (*string, error) {
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
		return nil, errors.New("attribute 'working_directory' must be a string")
	}
}

func (config *StepConfig) parseWorkingDirectory() (*string, error) {
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

func (config *StepConfig) unwrapIgnoreId(acto *actoparser.Acto) (*bool, error) {
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

		return config.unwrapIgnoreId(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'ignore_id' must be a bool")
	}
}

func (config *StepConfig) parseIgnoreId() (*bool, error) {
	acto := actoparser.NewActo(config.IgnoreId)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapIgnoreId(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *StepConfig) unwrapRun(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapRun(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'run' must be a string or a HEREDOC")
	}
}

func (config *StepConfig) parseRun() (*string, error) {
	acto := actoparser.NewActo(config.Run)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapRun(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *StepConfig) parseUses() (*string, error) {
	value, err := config.Uses.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in uses: %w", err)
	}

	return value, nil
}

func (config *StepConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'name' must be a string")
	}
}

func (config *StepConfig) parseName() (*string, error) {
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

func (config *StepConfig) unwrapIf(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapIf(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'if' must be a string or bool")
	}
}

func (config *StepConfig) parseIf() (*string, error) {
	acto := actoparser.NewActo(config.If)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapIf(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *StepConfig) Parse() (*Step, error) {
	if config == nil {
		return nil, nil
	}

	if config.Identifier == "" {
		return nil, fmt.Errorf("error in step: no identifier, %w", actoerrors.ErrOpenIssue)
	}

	step := Step{
		Identifier: config.Identifier,
	}

	ignoreId, err := config.parseIgnoreId()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if ignoreId == nil || !*ignoreId {
		step.Id = &step.Identifier
	}

	ifVal, err := config.parseIf()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if ifVal != nil {
		step.If = ifVal
	}

	name, err := config.parseName()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if name != nil {
		step.Name = name
	}

	uses, err := config.parseUses()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if uses != nil {
		step.Uses = uses
	}

	run, err := config.parseRun()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if run != nil {
		step.Run = run
	}

	if uses != nil && run != nil {
		return nil, fmt.Errorf("error in step '%s': only 'uses' or 'run', not both, %w", step.Identifier, actoerrors.ErrOpenIssue)
	}

	workingDirectory, err := config.parseWorkingDirectory()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if workingDirectory != nil {
		step.WorkingDirectory = workingDirectory
	}

	if uses != nil && workingDirectory != nil {
		return nil, fmt.Errorf("error in step '%s': 'working_directory' not allowed with 'uses', %w", step.Identifier, actoerrors.ErrOpenIssue)
	}

	shell, err := config.parseShell()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if shell != nil {
		step.Shell = shell
	}

	if uses != nil && shell != nil {
		return nil, fmt.Errorf("error in step '%s': 'shell' not allowed with 'uses', %w", step.Identifier, actoerrors.ErrOpenIssue)
	}

	withs, err := config.parseWith()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if withs != nil {
		step.With = withs
	}

	if withs != nil && uses == nil {
		return nil, fmt.Errorf("error in step '%s': can only have 'with' when 'uses' is set, %w", step.Identifier, actoerrors.ErrOpenIssue)
	}

	envs, err := config.parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if envs != nil {
		step.Env = envs
	}

	continueOnError, err := config.parseContinueOnError()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if continueOnError != nil {
		step.ContinueOnError = continueOnError
	}

	timeoutMinutes, err := config.parseTimeoutMinutes()
	if err != nil {
		return nil, fmt.Errorf("error in step '%s': %w, %w", step.Identifier, err, actoerrors.ErrOpenIssue)
	}

	if timeoutMinutes != nil {
		step.TimeoutMinutes = timeoutMinutes
	}

	return &step, nil
}

func (config *StepsConfig) Parse() (Steps, error) {
	steps := Steps{}

	for _, step := range *config {
		parsedStep, err := step.Parse()
		if err != nil {
			return Steps{}, err
		}

		if steps[parsedStep.Identifier] != nil {
			return Steps{}, fmt.Errorf("error in step '%s': already defined, %w", parsedStep.Identifier, actoerrors.ErrOpenIssue)
		}

		steps[parsedStep.Identifier] = parsedStep
	}

	return steps, nil
}

func (step *Step) Decode(body *hclwrite.Body, attr string) error {
	if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
		body.AppendNewline()
	}

	stepBlock := body.AppendNewBlock(attr, []string{step.Identifier})
	stepBody := stepBlock.Body()

	if step.Id != nil {
		idAttr, err := actoparser.GetHclTag(*step, "Id")
		if err != nil {
			return err
		}

		stepBody.SetAttributeValue(idAttr, cty.StringVal(*step.Id))
	}

	if step.If != nil {
		ifAttr, err := actoparser.GetHclTag(*step, "If")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue(ifAttr, cty.StringVal(*step.If))
	}

	if step.Name != nil {
		nameAttr, err := actoparser.GetHclTag(*step, "Name")
		if err != nil {
			return err
		}

		stepBody.SetAttributeValue(nameAttr, cty.StringVal(*step.Name))
	}

	if step.Uses != nil {
		usesAttr, err := actoparser.GetHclTag(*step, "Uses")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue(usesAttr, cty.StringVal(*step.Uses))
	}

	if step.Run != nil {
		runAttr, err := actoparser.GetHclTag(*step, "Run")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue(runAttr, cty.StringVal(*step.Run))
	}

	if step.WorkingDirectory != nil {
		workingDirectoryAttr, err := actoparser.GetHclTag(*step, "WorkingDirectory")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue(workingDirectoryAttr, cty.StringVal(*step.WorkingDirectory))
	}

	if step.Shell != nil {
		shellAttr, err := actoparser.GetHclTag(*step, "Shell")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue(shellAttr, cty.StringVal(*step.Shell))
	}

	if step.With != nil {
		withAttr, err := actoparser.GetHclTag(*step, "With")
		if err != nil {
			return err
		}

		for key, value := range *step.With {
			if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
				stepBody.AppendNewline()
			}

			withBlock := stepBody.AppendNewBlock(withAttr, nil)
			withBody := withBlock.Body()

			withBody.SetAttributeValue("name", cty.StringVal(key))

			switch v := value.(type) {
			case string:
				withBody.SetAttributeValue("value", cty.StringVal(v))
			case bool:
				withBody.SetAttributeValue("value", cty.BoolVal(v))
			case uint64:
				withBody.SetAttributeValue("value", cty.NumberUIntVal(v))
			case int64:
				withBody.SetAttributeValue("value", cty.NumberIntVal(v))
			case float64:
				withBody.SetAttributeValue("value", cty.NumberFloatVal(v))
			default:
				return errors.New("unkown dealt type")
			}
		}
	}

	if step.Env != nil {
		for name, env := range *step.Env {
			envAttr, err := actoparser.GetHclTag(*step, "Env")
			if err != nil {
				return err
			}

			if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
				stepBody.AppendNewline()
			}

			envBlock := stepBody.AppendNewBlock(envAttr, nil)

			envBody := envBlock.Body()
			envBody.SetAttributeValue("name", cty.StringVal(name))

			switch e := env.(type) {
			case string:
				envBody.SetAttributeValue("value", cty.StringVal(e))
			}
		}
	}

	if step.ContinueOnError != nil {
		attr, err := actoparser.GetHclTag(*step, "ContinueOnError")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 {
			stepBody.AppendNewline()
		}

		switch v := step.ContinueOnError.(type) {
		case string:
			stepBody.SetAttributeValue(attr, cty.StringVal(v))
		case bool:
			stepBody.SetAttributeValue(attr, cty.BoolVal(v))
		default:
			return errors.New("unkown dealt type")
		}
	}

	if step.TimeoutMinutes != nil {
		attr, err := actoparser.GetHclTag(*step, "TimeoutMinutes")
		if err != nil {
			return err
		}

		if len(stepBody.Blocks()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue(attr, cty.NumberUIntVal(*step.TimeoutMinutes))
	}

	return nil
}
