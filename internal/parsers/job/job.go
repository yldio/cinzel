// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type Job struct {
	Id       string   `yaml:"-"`
	Services Services `yaml:"services,omitempty"`
}
type Jobs map[string]any

type JobConfig struct {
	Id       string         `hcl:"id,label"` // check IdConfig and [issue](https://github.com/hashicorp/hcl/issues/583)
	Services ServicesConfig `hcl:"service,block"`
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

	services, err := config.Services.Parse()
	if err != nil {
		return Job{}, err
	}

	if len(services) != 0 {
		job.Services = services
	}

	return job, nil
}
