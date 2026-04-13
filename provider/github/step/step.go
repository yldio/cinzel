// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/provider/github/action"

	"github.com/zclconf/go-cty/cty"
)

// Steps maps step identifiers to their parsed Step representations.
type Steps map[string]Step

// Update sets the step's identifier to the given filename.
func (s *Step) Update(filename string) {
	s.Identifier = filename
}

// Step holds the parsed fields of a single GitHub Actions workflow step.
type Step struct {
	Identifier       string    `yaml:"-"`
	Id               cty.Value `yaml:"id,omitempty" hcl:"id"`
	If               cty.Value `yaml:"if,omitempty" hcl:"if"`
	Name             cty.Value `yaml:"name,omitempty" hcl:"name"`
	Uses             cty.Value `yaml:"uses,omitempty" hcl:"uses"`
	UsesComment      string    `yaml:"-"`
	Run              cty.Value `yaml:"run,omitempty" hcl:"run"`
	WorkingDirectory cty.Value `yaml:"working-directory,omitempty" hcl:"working_directory"`
	Shell            cty.Value `yaml:"shell,omitempty" hcl:"shell"`
	With             cty.Value `yaml:"with,omitempty" hcl:"with"`
	Env              cty.Value `yaml:"env,omitempty" hcl:"env"`
	ContinueOnError  cty.Value `yaml:"continue-on-error,omitempty" hcl:"continue_on_error"`
	TimeoutMinutes   cty.Value `yaml:"timeout-minutes,omitempty" hcl:"timeout_minutes"`
}

// StepListConfig is a slice of StepConfig decoded from HCL step blocks.
type StepListConfig []StepConfig

// StepConfig represents the HCL configuration for a single step block.
type StepConfig struct {
	Identifier       string                `hcl:"id,label"`
	Id               hcl.Expression        `hcl:"id,attr"`
	IgnoreId         hcl.Expression        `hcl:"ignore_id,attr"`
	If               hcl.Expression        `hcl:"if,attr"`
	Name             hcl.Expression        `hcl:"name,attr"`
	Uses             action.UsesListConfig `hcl:"uses,block"`
	Run              hcl.Expression        `hcl:"run,attr"`
	WorkingDirectory hcl.Expression        `hcl:"working_directory,attr"`
	Shell            hcl.Expression        `hcl:"shell,attr"`
	With             action.WithListConfig `hcl:"with,block"`
	Env              action.EnvListConfig  `hcl:"env,block"`
	ContinueOnError  hcl.Expression        `hcl:"continue_on_error,attr"`
	TimeoutMinutes   hcl.Expression        `hcl:"timeout_minutes,attr"`
}
