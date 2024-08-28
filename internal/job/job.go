// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/step"
)

type Jobs []Job

type Job struct {
	Id              string              `yaml:"-"`
	Name            *string             `yaml:"name,omitempty"`
	Permissions     *action.Permissions `yaml:"permissions,omitempty"`
	Needs           *[]string           `yaml:"needs,omitempty"`
	If              *string             `yaml:"id,omitempty"`
	RunsOn          *any                `yaml:"runs-on,omitempty"`
	Environment     *any                `yaml:"environment,omitempty"`
	Concurrency     *action.Concurrency `yaml:"concurrency,omitempty"`
	Outputs         *map[string]any     `yaml:"outputs,omitempty"`
	Env             *action.Env         `yaml:"env,omitempty"`
	Defaults        *action.Defaults    `yaml:"defaults,omitempty"`
	Steps           step.Steps          `yaml:"steps"`
	StepsIds        []string            `yaml:"-"`
	TimeoutMinutes  *uint16             `yaml:"timeout-minutes,omitempty"`
	Strategy        *action.Strategy    `yaml:"strategy,omitempty"`
	ContinueOnError *bool               `yaml:"continue-on-error,omitempty"`
	Container       *action.Container   `yaml:"container,omitempty"`
	Services        *action.Services    `yaml:"services,omitempty"`
	Uses            *string             `yaml:"uses,omitempty"`
	With            *map[string]any     `yaml:"with,omitempty"`
	Secrets         *any                `yaml:"secrets,omitempty"`
}

type JobsConfig []JobConfig

type JobConfig struct {
	Id              string                        `hcl:"id,label"`
	Name            *action.NameConfig            `hcl:"name,attr"`
	Permissions     *action.PermissionsConfig     `hcl:"permissions,block"`
	Needs           hcl.Expression                `hcl:"needs,attr"`
	If              *action.IfConfig              `hcl:"if,attr"`
	RunsOn          action.RunsOnConfig           `hcl:"runs,block"`
	Environment     *action.EnvironmentConfig     `hcl:"environment,block"`
	Concurrency     *action.ConcurrencyConfig     `hcl:"concurrency,block"`
	Outputs         action.OutputsConfig          `hcl:"output,block"` // action.OutputsConfig -> []*OutputConfig
	Env             *action.EnvConfig             `hcl:"env,block"`
	Defaults        *action.DefaultsConfig        `hcl:"defaults,block"`
	Steps           hcl.Expression                `hcl:"steps,attr"`
	TimeoutMinutes  *action.TimeoutMinutesConfig  `hcl:"timeout_minutes,attr"`
	Strategy        *action.StrategyConfig        `hcl:"strategy,block"`
	ContinueOnError *action.ContinueOnErrorConfig `hcl:"continue_on_error,attr"`
	Container       *action.ContainerConfig       `hcl:"container,block"`
	Services        action.ServicesConfig         `hcl:"service,block"` // action.ServicesConfig -> []*ServiceConfig
	Uses            *action.UsesConfig            `hcl:"uses,attr"`
	With            action.WithsConfig            `hcl:"with,block"` // action.WithsConfig -> []*WithConfig
	Secrets         *action.SecretsInheritConfig  `hcl:"secrets,attr"`
	Secret          action.SecretsConfig          `hcl:"secret,block"` // action.SecretsConfig -> []*SecretConfig
}

func (config *JobConfig) parseId(job *Job) error {
	if config.Id == "" {
		return nil
	}

	job.Id = config.Id

	return nil
}

func (config *JobConfig) parseName(job *Job) error {
	if config.Name == nil {
		return nil
	}

	if *config.Name == "" {
		return errors.New("name must be a non empty string")
	}
	name := string(*config.Name)

	job.Name = &name

	return nil
}

func (config *JobConfig) parsePermissions(job *Job) error {
	permissions, err := config.Permissions.Parse()
	if err != nil {
		return err
	}

	if permissions == (action.Permissions{}) {
		return nil
	}

	job.Permissions = &permissions

	return nil
}

func (config *JobConfig) parseNeeds(job *Job) error {
	if config.Needs == nil {
		return nil
	}

	val, diags := config.Needs.Value(nil)
	if diags.HasErrors() {
		// return errors.New(diags[0].Detail)
	}
	if val.IsNull() {
		return nil
	}

	exprs, diags := hcl.ExprList(config.Needs)
	if diags.HasErrors() {
		return errors.New(diags[0].Detail)
	}

	jobsIds := []string{}

	for _, expr := range exprs {
		traversal, diags := hcl.AbsTraversalForExpr(expr)
		if diags.HasErrors() {
			return errors.New(diags[0].Detail)
		}

		for _, traverser := range traversal {
			switch tJob := traverser.(type) {
			case hcl.TraverseRoot:
				if tJob.Name != "job" {
					return errors.New("needs require a job relationship only")
				}
			case hcl.TraverseAttr:
				jobsIds = append(jobsIds, tJob.Name)
			}
		}
	}

	if len(jobsIds) == 0 {
		return nil
	}

	job.Needs = &jobsIds

	return nil
}

func (config *JobConfig) parseIf(job *Job) error {
	iF, err := config.If.Parse()
	if err != nil {
		return err
	}

	if iF == "" {
		return nil
	}

	job.If = &iF

	return nil
}

func (config *JobConfig) parseRunsOn(job *Job) error {
	runsOn, err := config.RunsOn.Parse()
	if err != nil {
		return err
	}

	if runsOn == nil || runsOn == "" {
		return nil
	}

	job.RunsOn = &runsOn

	return nil
}

func (config *JobConfig) parseEnvironment(job *Job) error {
	environment, err := config.Environment.Parse()
	if err != nil {
		return err
	}

	if environment == nil {
		return nil
	}

	job.Environment = &environment

	return nil
}

func (config *JobConfig) parseConcurrency(job *Job) error {
	concurrency, err := config.Concurrency.Parse()
	if err != nil {
		return err
	}

	if concurrency == (action.Concurrency{}) {
		return nil
	}

	job.Concurrency = &concurrency

	return nil
}

func (config *JobConfig) parseOutputs(job *Job) error {
	outputs, err := config.Outputs.Parse()
	if err != nil {
		return err
	}

	if outputs == nil {
		return nil
	}

	job.Outputs = &outputs

	return nil
}

func (config *JobConfig) parseEnv(job *Job) error {
	env, err := config.Env.Parse()
	if err != nil {
		return err
	}

	if env == nil {
		return nil
	}

	job.Env = &env

	return nil
}

func (config *JobConfig) parseDefaults(job *Job) error {
	defaults, err := config.Defaults.Parse()
	if err != nil {
		return err
	}

	if defaults == (action.Defaults{}) {
		return nil
	}

	job.Defaults = &defaults

	return nil
}

func (config *JobConfig) parseSteps(job *Job) error {
	if config.Steps == nil {
		return nil
	}

	val, diags := config.Steps.Value(nil)
	if diags.HasErrors() {
		// return errors.New(diags[0].Detail)
	}
	if val.IsNull() {
		return nil
	}

	exprs, diags := hcl.ExprList(config.Steps)
	if diags.HasErrors() {
		return errors.New(diags[0].Detail)
	}

	stepsIds := []string{}

	for _, expr := range exprs {
		traversal, diags := hcl.AbsTraversalForExpr(expr)
		if diags.HasErrors() {
			return errors.New(diags[0].Detail)
		}

		for _, traverser := range traversal {
			switch tStep := traverser.(type) {
			case hcl.TraverseRoot:
				if tStep.Name != "step" {
					return errors.New("steps require a step relationship only")
				}
			case hcl.TraverseAttr:
				stepsIds = append(stepsIds, tStep.Name)
			}
		}
	}

	if len(stepsIds) == 0 {
		return errors.New("requires at least one step")
	}

	job.StepsIds = stepsIds

	return nil
}

func (config *JobConfig) parseTimeoutMinutes(job *Job) error {
	timeoutMinutes, err := config.TimeoutMinutes.Parse()
	if err != nil {
		return err
	}

	if timeoutMinutes == nil {
		return nil
	}

	job.TimeoutMinutes = timeoutMinutes

	return nil
}

func (config *JobConfig) parseContainer(job *Job) error {
	container, err := config.Container.Parse()
	if err != nil {
		return err
	}

	if container.IsNill() {
		return nil
	}

	job.Container = &container

	return nil
}

func (config *JobConfig) parseStrategy(job *Job) error {
	strategy, err := config.Strategy.Parse()
	if err != nil {
		return err
	}

	if strategy.Matrix == nil {
		return nil
	}

	job.Strategy = &strategy

	return nil
}

func (config *JobConfig) parseContinueOnError(job *Job) error {
	continueOnError, err := config.ContinueOnError.Parse()
	if err != nil {
		return err
	}

	if continueOnError == nil {
		return nil
	}

	job.ContinueOnError = continueOnError

	return nil
}

func (config *JobConfig) parseServices(job *Job) error {
	services, err := config.Services.Parse()
	if err != nil {
		return err
	}

	if services.IsNill() {
		return nil
	}

	job.Services = &services

	return nil
}

func (config *JobConfig) parseUses(job *Job) error {
	uses, err := config.Uses.Parse()
	if err != nil {
		return err
	}

	if uses == "" {
		return nil
	}

	job.Uses = &uses

	return nil
}

func (config *JobConfig) parseWith(job *Job) error {
	with, err := config.With.Parse()
	if err != nil {
		return err
	}

	if with == nil {
		return nil
	}

	job.With = &with

	return nil
}

func (config *JobConfig) parseSecrets(job *Job) error {
	secretsInherit, err := config.Secrets.Parse()
	if err != nil {
		return err
	}

	secretsList, err := config.Secret.Parse()
	if err != nil {
		return err
	}

	var secrets any

	inheritNil := secretsInherit.IsNill()
	secretsNil := secretsList.IsNill()

	if !inheritNil && !secretsNil {
		return errors.New("only `secrets` blocks or one single `secret` attribute is allowed")
	}

	if !inheritNil {
		secrets = &secretsInherit
	} else if !secretsNil {
		secrets = &secretsList
	}

	if secrets != nil {
		job.Secrets = &secrets
	}

	return nil
}

func (config *JobConfig) Parse() (Job, error) {
	if config == nil {
		return Job{}, nil
	}

	job := Job{
		Id: config.Id,
	}

	if err := config.parseId(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseName(&job); err != nil {
		return Job{}, err
	}

	if err := config.parsePermissions(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseNeeds(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseIf(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseRunsOn(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseEnvironment(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseConcurrency(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseOutputs(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseEnv(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseDefaults(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseSteps(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseTimeoutMinutes(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseStrategy(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseContinueOnError(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseContainer(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseServices(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseUses(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseUses(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseWith(&job); err != nil {
		return Job{}, err
	}

	if err := config.parseSecrets(&job); err != nil {
		return Job{}, err
	}

	return job, nil
}

func (config *JobsConfig) Parse() (Jobs, error) {
	jobs := Jobs{}

	for _, job := range *config {
		parsedjob, err := job.Parse()
		if err != nil {
			return Jobs{}, err
		}

		jobs = append(jobs, parsedjob)
	}

	return jobs, nil
}

func (jobs *Jobs) PostParseNeeds(parsedJobs Jobs) error {
	// TODO: validate job_1 exists
	return nil
}

func (jobs *Jobs) PostParseSteps(parsedSteps step.Steps) error {
	for idx, job := range *jobs {
		if len(job.StepsIds) == 0 {
			return fmt.Errorf("job %s requires at least one step", job.Id)
		}

		stepsList := step.Steps{}

		for _, stepId := range job.StepsIds {
			for _, parsedStep := range parsedSteps {
				if parsedStep.Id != stepId {
					continue
				}

				stepsList = append(stepsList, parsedStep)
			}
		}

		if len(stepsList) == 0 && len(parsedSteps) > 0 {
			return errors.New("some steps do not exist")
		}

		job.Steps = stepsList

		(*jobs)[idx] = job
	}

	return nil
}
