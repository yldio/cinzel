// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package github

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/provider/github/job"
	"github.com/yldio/acto/provider/github/step"
	"github.com/yldio/acto/provider/github/variable"
	"github.com/yldio/acto/provider/github/workflow"
)

const defaultOutputDirectory = ".github/workflows"

type ConfigHCL struct {
	Steps     step.StepsConfig         `hcl:"step,block"`
	Jobs      job.JobsConfig           `hcl:"job,block"`
	Workflows workflow.WorkflowsConfig `hcl:"workflow,block"`
	Variables variable.VariablesConfig `hcl:"variable,block"`
}

type GitHub struct {
	configHCL   ConfigHCL
	bodyHCL     hcl.Body
	name        string
	description string
}

func (p *GitHub) DefaultOutputDirectory() string {
	return defaultOutputDirectory
}

func (p *GitHub) GetName() string {
	return p.name
}

func (p *GitHub) GetDescription() string {
	return p.description
}

func New() *GitHub {
	return &GitHub{
		name:        "github",
		description: "",
	}
}
