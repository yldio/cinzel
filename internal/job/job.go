// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package job

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/step"
	"github.com/yldio/acto/internal/variables"
)

type Jobs map[string]*Job

type Job struct {
	Id              string              `yaml:"-"`
	Name            *string             `yaml:"name,omitempty"`
	Permissions     *action.Permissions `yaml:"permissions,omitempty"`
	Needs           *[]string           `yaml:"needs,omitempty"`
	If              *string             `yaml:"if,omitempty"`
	RunsOn          any                 `yaml:"runs-on,omitempty"`
	Environment     any                 `yaml:"environment,omitempty"`
	Concurrency     *action.Concurrency `yaml:"concurrency,omitempty"`
	Outputs         *action.Outputs     `yaml:"outputs,omitempty"`
	Env             *action.Envs        `yaml:"env,omitempty"`
	Defaults        *action.Defaults    `yaml:"defaults,omitempty"`
	Steps           []*step.Step        `yaml:"steps,omitempty"`
	StepsIds        []string            `yaml:"-"`
	TimeoutMinutes  *uint64             `yaml:"timeout-minutes,omitempty"`
	Strategy        *action.Strategy    `yaml:"strategy,omitempty"`
	ContinueOnError any                 `yaml:"continue-on-error,omitempty"`
	Container       *action.Container   `yaml:"container,omitempty"`
	Services        *action.Services    `yaml:"services,omitempty"`
	Uses            *string             `yaml:"uses,omitempty"`
	With            *map[string]any     `yaml:"with,omitempty"`
	Secrets         any                 `yaml:"secrets,omitempty"`
}

type JobsConfig []JobConfig

type JobConfig struct {
	Identifier      string                    `hcl:"id,label"`
	Name            hcl.Expression            `hcl:"name,attr"`
	Permissions     *action.PermissionsConfig `hcl:"permissions,block"`
	Needs           hcl.Expression            `hcl:"needs,attr"`
	If              hcl.Expression            `hcl:"if,attr"`
	RunsOn          *action.RunsOnConfig      `hcl:"runs_on,block"`
	Environment     *action.EnvironmentConfig `hcl:"environment,block"`
	Concurrency     *action.ConcurrencyConfig `hcl:"concurrency,block"`
	Outputs         action.OutputsConfig      `hcl:"output,block"`
	Env             action.EnvsConfig         `hcl:"env,block"`
	Defaults        *action.DefaultsConfig    `hcl:"defaults,block"`
	Steps           hcl.Expression            `hcl:"steps,attr"`
	TimeoutMinutes  hcl.Expression            `hcl:"timeout_minutes,attr"`
	Strategy        *action.StrategyConfig    `hcl:"strategy,block"`
	ContinueOnError hcl.Expression            `hcl:"continue_on_error,attr"`
	Container       *action.ContainerConfig   `hcl:"container,block"`
	Services        action.ServicesConfig     `hcl:"service,block"`
	Uses            *action.UsesConfig        `hcl:"uses,block"`
	With            action.WithsConfig        `hcl:"with,block"`
	Secrets         action.SecretsConfig      `hcl:"secret,block"`
	SecretsInherit  hcl.Expression            `hcl:"secrets,attr"`
}

const Inherit = "inherit"

func (config *JobConfig) unwrapSecretsInherit(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		if resultValue != Inherit {
			return nil, fmt.Errorf("attribute 'secrets' must be the hardcoded string '%s'", Inherit)
		}
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapSecretsInherit(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, fmt.Errorf("attribute 'secrets' must be the hardcoded string '%s'", Inherit)
	}
}

func (config *JobConfig) parseSecrets() (any, error) {
	acto := actoparser.NewActo(config.SecretsInherit)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	secretsInherit, err := config.unwrapSecretsInherit(acto)
	if err != nil {
		return nil, err
	}

	secrets, err := config.Secrets.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in secrets: %w", err)
	}

	if secrets != nil && len(*secrets) > 0 && secretsInherit != nil {
		return nil, fmt.Errorf("error in secrets: can only have 'secrets' inherit or a set of secrets")
	}

	if secretsInherit != nil {
		return secretsInherit, nil
	}

	if secrets == nil {
		return nil, nil
	}

	return secrets, nil
}

func (config *JobConfig) parseWith() (*map[string]any, error) {
	value, err := config.With.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in with: %w", err)
	}

	return value, nil
}

func (config *JobConfig) parseUses() (*string, error) {
	value, err := config.Uses.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in uses: %w", err)
	}

	return value, nil
}

func (config *JobConfig) parseServices() (*action.Services, error) {
	services := make(action.Services)

	for _, service := range config.Services {
		svc, err := service.Parse()
		if err != nil {
			return nil, fmt.Errorf("error in service: %w", err)
		}

		if services[svc.Name] != nil {
			return nil, fmt.Errorf("error in service: '%s' already defined ", svc.Name)
		}

		services[svc.Name] = svc
	}

	if len(services) == 0 {
		return nil, nil
	}

	return &services, nil
}

func (config *JobConfig) parseContainer() (*action.Container, error) {
	container, err := config.Container.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in container: %w", err)
	}

	return container, nil
}

func (config *JobConfig) unwrapContinueOnError(acto *actoparser.Acto) (any, error) {
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

func (config *JobConfig) parseContinueOnError() (any, error) {
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

func (config *JobConfig) parseStrategy() (*action.Strategy, error) {
	strategy, err := config.Strategy.Parse()
	if err != nil {
		return nil, err
	}

	return strategy, nil
}

func (config *JobConfig) unwrapTimeoutMinutes(acto *actoparser.Acto) (*uint64, error) {
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

func (config *JobConfig) parseTimeoutMinutes() (*uint64, error) {
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

func (config *JobConfig) unwrapStepsIds(acto *actoparser.Acto) (*[]string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []actoparser.ActoVariableRef:
		list := []string{}
		for _, stepRef := range resultValue {
			if stepRef.Name != "step" {
				return nil, errors.New("invalid step reference, should be step.<step-identifier>")
			}

			list = append(list, stepRef.Attr)
		}

		return &list, nil
	default:
		return nil, errors.New("attribute 'Steps' must be a list of steps relation")
	}
}

func (config *JobConfig) parseStepsIds() (*[]string, error) {
	acto := actoparser.NewActo(config.Steps)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	stepsIds, err := config.unwrapStepsIds(acto)
	if err != nil {
		return nil, err
	}

	if stepsIds != nil && len(*stepsIds) == 0 {
		return nil, errors.New("attribute 'steps' cannot be empty")
	}

	return stepsIds, nil
}

func (config *JobConfig) parseDefaults() (*action.Defaults, error) {
	defaults, err := config.Defaults.Parse()
	if err != nil {
		return nil, err
	}

	return defaults, nil
}

func (config *JobConfig) parseEnvs() (*action.Envs, error) {
	env, err := config.Env.Parse()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (config *JobConfig) parseOutputs() (*action.Outputs, error) {
	outputs, err := config.Outputs.Parse()
	if err != nil {
		return nil, err
	}

	if outputs == nil || len(*outputs) == 0 {
		return nil, nil
	}

	return outputs, nil
}

func (config *JobConfig) parseConcurrency() (*action.Concurrency, error) {
	concurrency, err := config.Concurrency.Parse()
	if err != nil {
		return nil, err
	}

	if concurrency == nil {
		return nil, nil
	}

	return concurrency, nil
}

func (config *JobConfig) parseEnvironment() (any, error) {
	environment, err := config.Environment.Parse()
	if err != nil {
		return nil, err
	}

	return environment, nil
}

func (config *JobConfig) unwrapIf(acto *actoparser.Acto) (*string, error) {
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

func (config *JobConfig) unwrapNeeds(acto *actoparser.Acto) (*[]string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []actoparser.ActoVariableRef:
		list := []string{}
		for _, jobRef := range resultValue {
			if jobRef.Name != "job" {
				return nil, errors.New("invalid job reference, should be job.<job-identifier>")
			}

			list = append(list, jobRef.Attr)
		}

		return &list, nil
	default:
		return nil, errors.New("attribute 'needs' must be a list of jobs relation")
	}
}

func (config *JobConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
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

func (config *JobConfig) parseName() (*string, error) {
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

func (config *JobConfig) parsePermissions() (*action.Permissions, error) {
	permissions, err := config.Permissions.Parse()
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (config *JobConfig) parseNeeds() (*[]string, error) {
	acto := actoparser.NewActo(config.Needs)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapNeeds(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *JobConfig) parseIf() (*string, error) {
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

func (config *JobConfig) parseRunsOn() (any, error) {
	runsOn, err := config.RunsOn.Parse()
	if err != nil {
		return nil, err
	}

	return runsOn, nil
}

func (config *JobConfig) Parse() (*Job, error) {
	if config == nil {
		return nil, nil
	}

	if config.Identifier == "" {
		return nil, fmt.Errorf("error in job: no identifier, %w", actoerrors.ErrOpenIssue)
	}

	job := Job{
		Id: config.Identifier,
	}

	name, err := config.parseName()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if name != nil {
		job.Name = name
	}

	permissions, err := config.parsePermissions()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if permissions != nil {
		job.Permissions = permissions
	}

	needs, err := config.parseNeeds()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if needs != nil {
		job.Needs = needs
	}

	ifVal, err := config.parseIf()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if ifVal != nil {
		job.If = ifVal
	}

	runsOn, err := config.parseRunsOn()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if runsOn != nil {
		job.RunsOn = runsOn
	}

	environment, err := config.parseEnvironment()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if environment != nil {
		job.Environment = environment
	}

	concurrency, err := config.parseConcurrency()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if concurrency != nil {
		job.Concurrency = concurrency
	}

	outputs, err := config.parseOutputs()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if outputs != nil {
		job.Outputs = outputs
	}

	envs, err := config.parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if envs != nil {
		job.Env = envs
	}

	defaults, err := config.parseDefaults()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if defaults != nil {
		job.Defaults = defaults
	}

	stepsIds, err := config.parseStepsIds()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if stepsIds != nil {
		job.StepsIds = *stepsIds
	}

	timeoutMinutes, err := config.parseTimeoutMinutes()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if timeoutMinutes != nil {
		job.TimeoutMinutes = timeoutMinutes
	}

	strategy, err := config.parseStrategy()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if strategy != nil {
		job.Strategy = strategy
	}

	continueOnError, err := config.parseContinueOnError()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if continueOnError != nil {
		job.ContinueOnError = continueOnError
	}

	container, err := config.parseContainer()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if container != nil {
		job.Container = container
	}

	services, err := config.parseServices()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if services != nil {
		job.Services = services
	}

	uses, err := config.parseUses()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if uses != nil {
		job.Uses = uses
	}

	if runsOn != nil && uses != nil {
		return nil, fmt.Errorf("error in job '%s': can only have 'runs_on' or 'uses', not both, %w", job.Id, actoerrors.ErrOpenIssue)
	}

	withs, err := config.parseWith()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if withs != nil {
		job.With = withs
	}

	if withs != nil && uses == nil {
		return nil, fmt.Errorf("error in job '%s': can only have 'with' when 'uses' is set, %w", job.Id, actoerrors.ErrOpenIssue)
	}

	secrets, err := config.parseSecrets()
	if err != nil {
		return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
	}

	if secrets != nil {
		job.Secrets = secrets
	}

	if secrets != nil && uses == nil {
		return nil, fmt.Errorf("error in job '%s': can only have 'secret' when 'uses' is set, %w", job.Id, actoerrors.ErrOpenIssue)
	}

	return &job, nil
}

func (config *JobsConfig) Parse() (Jobs, error) {
	jobs := Jobs{}

	for _, job := range *config {
		parsedJob, err := job.Parse()
		if err != nil {
			return Jobs{}, err
		}

		if jobs[parsedJob.Id] != nil {
			return Jobs{}, fmt.Errorf("error in job '%s': already defined, %w", parsedJob.Id, actoerrors.ErrOpenIssue)
		}

		jobs[parsedJob.Id] = parsedJob
	}

	return jobs, nil
}
