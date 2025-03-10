// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/provider/github/step"
)

type hclJobBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type hclWorkflowBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type hclActionBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type parseConfig struct {
	Variables hclparser.VariablesConfig `hcl:"variable,block"`
	Steps     step.StepListConfig       `hcl:"step,block"`
	Jobs      []hclJobBlock             `hcl:"job,block"`
	Workflows []hclWorkflowBlock        `hcl:"workflow,block"`
	Actions   []hclActionBlock          `hcl:"action,block"`
}
