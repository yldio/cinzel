// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package github

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/provider/github/job"
	"github.com/yldio/acto/provider/github/step"
	"github.com/yldio/acto/provider/github/variable"
	"github.com/yldio/acto/provider/github/workflow"
)

type HclGitHubConfig struct {
	Steps     step.StepsConfig         `hcl:"step,block"`
	Jobs      job.JobsConfig           `hcl:"job,block"`
	Workflows workflow.WorkflowsConfig `hcl:"workflow,block"`
	Variables variable.VariablesConfig `hcl:"variable,block"`
}

func (p *GitHub) Do() (workflow.Workflows, error) {
	diags := gohcl.DecodeBody(p.bodyHCL, nil, &p.configHCL)
	if diags.HasErrors() {
		return workflow.Workflows{}, actoerrors.ProcessHCLDiags(diags)
	}

	return p.parse()
}

func (p *GitHub) parse() (workflow.Workflows, error) {
	if err := p.configHCL.Variables.Parse(); err != nil {
		return workflow.Workflows{}, err
	}

	parsedSteps, err := p.configHCL.Steps.Parse()
	if err != nil {
		return workflow.Workflows{}, err
	}

	parsedJobs, err := p.configHCL.Jobs.Parse()
	if err != nil {
		return workflow.Workflows{}, err
	}

	for _, job := range parsedJobs {
		if job.Needs == nil {
			continue
		}

		for _, jobId := range *job.Needs {
			if parsedJobs[jobId] == nil {
				err := fmt.Errorf("cannot find job with '%s' identifier", jobId)
				return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
			}
		}
	}

	for _, job := range parsedJobs {
		for _, stepId := range job.StepsIds {
			if parsedSteps[stepId] == nil {
				err := fmt.Errorf("cannot find step with '%s' identifier", stepId)
				return nil, fmt.Errorf("error in job '%s': %w, %w", job.Id, err, actoerrors.ErrOpenIssue)
			}

			if job.Steps == nil {
				job.Steps = []*step.Step{}
			}

			job.Steps = append(job.Steps, parsedSteps[stepId])
		}
	}

	parsedWorkflows, err := p.configHCL.Workflows.Parse()
	if err != nil {
		return workflow.Workflows{}, err
	}

	for _, workflow := range parsedWorkflows {
		for _, jobId := range workflow.JobsIds {
			if parsedJobs[jobId] == nil {
				err := fmt.Errorf("cannot find job with '%s' identifier", jobId)
				return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
			}

			if workflow.Jobs == nil {
				workflow.Jobs = make(job.Jobs)
			}

			workflow.Jobs[jobId] = parsedJobs[jobId]
		}

	}

	return parsedWorkflows, nil
}
