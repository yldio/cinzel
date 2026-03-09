// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"github.com/hashicorp/hcl/v2"
)

type hclVariableBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type hclJobBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type hclWorkflowBlock struct {
	Body hcl.Body `hcl:",remain"`
}

type hclTemplateBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type parseConfig struct {
	Stages    []string           `hcl:"stages,optional"`
	Variables []hclVariableBlock `hcl:"variable,block"`
	Jobs      []hclJobBlock      `hcl:"job,block"`
	Workflow  []hclWorkflowBlock `hcl:"workflow,block"`
	Templates []hclTemplateBlock `hcl:"template,block"`
}
