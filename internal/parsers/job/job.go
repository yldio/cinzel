// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"errors"

	"github.com/yldio/atos/internal/parsers"
)

// order of properties matter when converting to Yaml
type Job struct {
	Id        string    `yaml:"-"`
	Container Container `yaml:"container,omitempty"`
	Services  Services  `yaml:"services,omitempty"`
	Secrets   any       `yaml:"secrets,omitempty"`
	Uses      Uses      `yaml:"uses,omitempty"`
	With      With      `yaml:"with,omitempty"`
	Strategy  Strategy  `yaml:"strategy,omitempty"`
}
type Jobs map[string]Job

type JobConfig struct {
	Id        string               `hcl:"id,label"` // check IdConfig and [issue](https://github.com/hashicorp/hcl/issues/583)
	Container ContainerConfig      `hcl:"container,block"`
	Services  ServicesConfig       `hcl:"service,block"`
	Secret    SecretsConfig        `hcl:"secret,block"`
	Secrets   SecretsInheritConfig `hcl:"secrets,attr"`
	With      WithConfig           `hcl:"with,block"`
	Uses      UsesConfig           `hcl:"uses,attr"`
	Strategy  StrategyConfig       `hcl:"strategy,block"`
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

	return job, nil
}
