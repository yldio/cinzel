// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"errors"

	"github.com/yldio/atos/internal/parsers"
)

// order of properties matter when converting to Yaml
type Job struct {
	Id              string      `yaml:"-"`
	Name            string      `yaml:"name,omitempty"`
	If              string      `yaml:"if,omitempty"`
	RunsOn          RunsOn      `yaml:"runs-on,omitempty"`
	Env             Env         `yaml:"env,omitempty"`
	Environment     Environment `yaml:"environment,omitempty"`
	Container       Container   `yaml:"container,omitempty"`
	Services        Services    `yaml:"services,omitempty"`
	Secrets         any         `yaml:"secrets,omitempty"`
	Uses            Uses        `yaml:"uses,omitempty"`
	With            With        `yaml:"with,omitempty"`
	Strategy        Strategy    `yaml:"strategy,omitempty"`
	ContinueOnError bool        `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes  uint16      `yaml:"timeout-minutes,omitempty"`
}
type Jobs map[string]Job

type JobConfig struct {
	Id              string                `hcl:"id,label"` // check IdConfig and [issue](https://github.com/hashicorp/hcl/issues/583)
	Name            NameConfig            `hcl:"name,attr"`
	If              IfConfig              `hcl:"if,attr"`
	Runs            RunsConfig            `hcl:"runs,block"`
	Env             EnvConfig             `hcl:"env,block"`
	Environment     EnvironmentConfig     `hcl:"environment,block"`
	Container       ContainerConfig       `hcl:"container,block"`
	Services        ServicesConfig        `hcl:"service,block"`
	Secret          SecretsConfig         `hcl:"secret,block"`
	Secrets         SecretsInheritConfig  `hcl:"secrets,attr"`
	With            WithConfig            `hcl:"with,block"`
	Uses            UsesConfig            `hcl:"uses,attr"`
	Strategy        StrategyConfig        `hcl:"strategy,block"`
	ContinueOnError ContinueOnErrorConfig `hcl:"continue_on_error,attr"`
	TimeoutMinutes  TimeoutMinutesConfig  `hcl:"timeout_minutes,attr"`
}

type JobsConfig []JobConfig

func (config *JobsConfig) Parse() (Jobs, error) {
	jobs := make(Jobs)
	for _, job := range *config {
		parsedJob, err := job.Parse()
		if err != nil {
			return Jobs{}, nil
		}

		jobs[parsedJob.Id] = parsedJob
	}
	return jobs, nil
}

func (config *JobConfig) Parse() (Job, error) {
	job := Job{
		Id: config.Id,
	}

	jobName, err := config.Name.Parse()
	if err != nil {
		return Job{}, err
	}

	if jobName != "" {
		job.Name = jobName
	}

	jobRuns, err := config.Runs.Parse()
	if err != nil {
		return Job{}, err
	}

	if jobRuns != "" {
		job.RunsOn = jobRuns
	}

	jobEnv, err := config.Env.Parse()
	if err != nil {
		return Job{}, err
	}

	if len(jobEnv) != 0 {
		job.Env = jobEnv
	}

	if config.Environment != (EnvironmentConfig{}) {
		jobEnvironment, err := config.Environment.Parse()
		if err != nil {
			return Job{}, err
		}

		if jobEnvironment != nil {
			job.Environment = jobEnvironment
		}
	}

	jobIf, err := config.If.Parse()
	if err != nil {
		return Job{}, err
	}

	if jobIf != "" {
		job.If = jobIf
	}

	jobUses, err := config.Uses.Parse()
	if err != nil {
		return Job{}, err
	}

	if jobUses != "" {
		job.Uses = jobUses
	}

	jobWith, err := config.With.Parse()
	if err != nil {
		return Job{}, err
	}

	if len(jobWith) != 0 {
		job.With = jobWith
	}

	jobContainer, err := config.Container.Parse()
	if err != nil {
		return Job{}, err
	}

	job.Container = jobContainer

	jobStrategy, err := config.Strategy.Parse()
	if err != nil {
		return Job{}, err
	}

	job.Strategy = jobStrategy

	jobServices, err := config.Services.Parse()
	if err != nil {
		return Job{}, err
	}

	if len(jobServices.Services) != 0 {
		job.Services = jobServices.Services
	}

	if config.Secret != nil && config.Secrets != "" {
		return Job{}, errors.New(parsers.ErrorJobSecretsRestriction)
	}

	if config.Secrets != "" {
		secrets, err := config.Secrets.Parse()
		if err != nil {
			return Job{}, err
		}

		if secrets != "" {
			job.Secrets = secrets
		}
	} else if config.Secret != nil {
		secrets, err := config.Secret.Parse()
		if err != nil {
			return Job{}, err
		}

		if len(secrets) != 0 {
			job.Secrets = secrets
		}
	}

	jobContinueOnError, err := config.ContinueOnError.Parse()
	if err != nil {
		return Job{}, err
	}

	job.ContinueOnError = jobContinueOnError

	jobTimeoutMinutes, err := config.TimeoutMinutes.Parse()
	if err != nil {
		return Job{}, err
	}
	job.TimeoutMinutes = jobTimeoutMinutes

	return job, nil
}
