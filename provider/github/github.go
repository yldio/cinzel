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
	configHCL          ConfigHCL
	bodyHCL            hcl.Body
	provider           string
	description        string
	parseDescription   string
	unparseDescription string
}

func (p *GitHub) DefaultOutputDirectory() string {
	return defaultOutputDirectory
}

func (p *GitHub) GetProviderName() string {
	return p.provider
}

func (p *GitHub) GetDescription() string {
	return p.description
}

func (p *GitHub) GetParseDescription() string {
	return p.parseDescription
}

func (p *GitHub) GetUnparseDescription() string {
	return p.unparseDescription
}

func New() *GitHub {
	return &GitHub{
		provider:           "github",
		description:        "MISSING DESCRIPTION",
		parseDescription:   "Convert HCL files describing workflows to GitHub's Actions Yaml files",
		unparseDescription: "Converts existing GitHub's Actions Yaml files into HCL",
	}
}
