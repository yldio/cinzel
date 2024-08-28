// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package hclparser

import (
	"errors"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/yldio/acto/internal/job"
	"github.com/yldio/acto/internal/step"
	"github.com/yldio/acto/internal/workflow"
	"github.com/yldio/acto/service/yamlparser"
)

type HclConfig struct {
	Steps     step.StepsConfig         `hcl:"step,block"`
	Jobs      job.JobsConfig           `hcl:"job,block"`
	Workflows workflow.WorkflowsConfig `hcl:"workflow,block"`
}

type HclParse struct {
	body   hcl.Body
	config HclConfig
}

func New(body hcl.Body) *HclParse {
	return &HclParse{
		body: body,
	}
}

func (parse *HclParse) Decode() error {
	diags := gohcl.DecodeBody(parse.body, nil, &parse.config)
	if diags.HasErrors() {
		return errors.New(diags[0].Detail)
	}

	return nil
}

func (parse *HclParse) Parse() (*yamlparser.Yaml, error) {
	parsedSteps, err := parse.config.Steps.Parse()
	if err != nil {
		return &yamlparser.Yaml{}, err
	}

	parsedJobs, err := parse.config.Jobs.Parse()
	if err != nil {
		return &yamlparser.Yaml{}, err
	}

	parsedWorkflows, err := parse.config.Workflows.Parse()
	if err != nil {
		return &yamlparser.Yaml{}, err
	}

	if err := parsedJobs.PostParseSteps(parsedSteps); err != nil {
		return &yamlparser.Yaml{}, err
	}

	if err := parsedJobs.PostParseNeeds(parsedJobs); err != nil {
		return &yamlparser.Yaml{}, err
	}

	if err := parsedWorkflows.PostParse(parsedJobs); err != nil {
		return &yamlparser.Yaml{}, err
	}

	yamlparser := yamlparser.New(parsedWorkflows)

	return yamlparser, nil
}
