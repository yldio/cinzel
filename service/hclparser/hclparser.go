// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package hclparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/job"
	"github.com/yldio/acto/internal/step"
	"github.com/yldio/acto/internal/variable"
	"github.com/yldio/acto/internal/workflow"
	"github.com/yldio/acto/service/yamlwriter"
)

type HclConfig struct {
	Steps     step.StepsConfig         `hcl:"step,block"`
	Jobs      job.JobsConfig           `hcl:"job,block"`
	Workflows workflow.WorkflowsConfig `hcl:"workflow,block"`
	Variables variable.VariablesConfig `hcl:"variable,block"`
}

type HclParser struct {
	body   hcl.Body
	config HclConfig
}

func New(body hcl.Body) *HclParser {
	return &HclParser{
		body: body,
	}
}

// Decode is a wrapper around the gohcl.DecodeBody function.
func (parse *HclParser) Decode() error {
	diags := gohcl.DecodeBody(parse.body, nil, &parse.config)
	if diags.HasErrors() {
		return actoerrors.ProcessHCLDiags(diags)
	}

	return nil
}

func (parse *HclParser) Parse() (*yamlwriter.Yaml, error) {
	err := parse.config.Variables.Parse()
	if err != nil {
		return &yamlwriter.Yaml{}, err
	}

	parsedSteps, err := parse.config.Steps.Parse()
	if err != nil {
		return &yamlwriter.Yaml{}, err
	}

	parsedJobs, err := parse.config.Jobs.Parse()
	if err != nil {
		return &yamlwriter.Yaml{}, err
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

	parsedWorkflows, err := parse.config.Workflows.Parse()
	if err != nil {
		return &yamlwriter.Yaml{}, err
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

	yamlwriter := yamlwriter.New(parsedWorkflows)

	return yamlwriter, nil
}
