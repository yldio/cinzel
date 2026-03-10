// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/provider/github/step"
)

type hclNamedBlock struct {
	Name  hcl.Expression `hcl:"name"`
	Value hcl.Expression `hcl:"value"`
}

type hclOnBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type hclUsesBlock struct {
	Action  hcl.Expression `hcl:"action,optional"`
	Version hcl.Expression `hcl:"version,optional"`
}

type hclServiceBlock struct {
	ID   string   `hcl:"id,label"`
	Body hcl.Body `hcl:",remain"`
}

type hclRunsOnBlock struct {
	Body hcl.Body `hcl:",remain"`
}

type hclStrategyBlock struct {
	Body hcl.Body `hcl:",remain"`
}

type hclGenericBlock struct {
	Body hcl.Body `hcl:",remain"`
}

type hclJobBlock struct {
	ID              string             `hcl:"id,label"`
	Name            hcl.Expression     `hcl:"name,optional"`
	If              hcl.Expression     `hcl:"if,optional"`
	Uses            hcl.Expression     `hcl:"uses,optional"`
	Steps           hcl.Expression     `hcl:"steps,optional"`
	DependsOn       hcl.Expression     `hcl:"depends_on,optional"`
	Secrets         hcl.Expression     `hcl:"secrets,optional"`
	ContinueOnError hcl.Expression     `hcl:"continue_on_error,optional"`
	TimeoutMinutes  hcl.Expression     `hcl:"timeout_minutes,optional"`
	UsesBlocks      []hclUsesBlock     `hcl:"uses,block"`
	WithBlocks      []hclNamedBlock    `hcl:"with,block"`
	EnvBlocks       []hclNamedBlock    `hcl:"env,block"`
	OutputBlocks    []hclNamedBlock    `hcl:"output,block"`
	SecretBlocks    []hclNamedBlock    `hcl:"secret,block"`
	ServiceBlocks   []hclServiceBlock  `hcl:"service,block"`
	RunsOnBlocks    []hclRunsOnBlock   `hcl:"runs_on,block"`
	StrategyBlocks  []hclStrategyBlock `hcl:"strategy,block"`
	Permissions     []hclGenericBlock  `hcl:"permissions,block"`
	Defaults        []hclGenericBlock  `hcl:"defaults,block"`
	Concurrency     []hclGenericBlock  `hcl:"concurrency,block"`
	Container       []hclGenericBlock  `hcl:"container,block"`
	Environment     []hclGenericBlock  `hcl:"environment,block"`
}

type hclWorkflowBlock struct {
	ID          string            `hcl:"id,label"`
	Filename    hcl.Expression    `hcl:"filename,optional"`
	Name        hcl.Expression    `hcl:"name,optional"`
	RunName     hcl.Expression    `hcl:"run_name,optional"`
	Jobs        hcl.Expression    `hcl:"jobs,optional"`
	Permissions hcl.Expression    `hcl:"permissions,optional"`
	Concurrency hcl.Expression    `hcl:"concurrency,optional"`
	On          []hclOnBlock      `hcl:"on,block"`
	Env         []hclNamedBlock   `hcl:"env,block"`
	PermBlocks  []hclGenericBlock `hcl:"permissions,block"`
	Defaults    []hclGenericBlock `hcl:"defaults,block"`
	ConcBlocks  []hclGenericBlock `hcl:"concurrency,block"`
}

type hclActionInputBlock struct {
	ID                 string         `hcl:"id,label"`
	Description        hcl.Expression `hcl:"description,optional"`
	Required           hcl.Expression `hcl:"required,optional"`
	Default            hcl.Expression `hcl:"default,optional"`
	DeprecationMessage hcl.Expression `hcl:"deprecation_message,optional"`
}

type hclActionOutputBlock struct {
	ID          string         `hcl:"id,label"`
	Description hcl.Expression `hcl:"description,optional"`
	Value       hcl.Expression `hcl:"value,optional"`
}

type hclActionRunsBlock struct {
	Using      hcl.Expression  `hcl:"using,optional"`
	Main       hcl.Expression  `hcl:"main,optional"`
	Pre        hcl.Expression  `hcl:"pre,optional"`
	PreIf      hcl.Expression  `hcl:"pre_if,optional"`
	Post       hcl.Expression  `hcl:"post,optional"`
	PostIf     hcl.Expression  `hcl:"post_if,optional"`
	Image      hcl.Expression  `hcl:"image,optional"`
	Args       hcl.Expression  `hcl:"args,optional"`
	Entrypoint hcl.Expression  `hcl:"entrypoint,optional"`
	Steps      hcl.Expression  `hcl:"steps,optional"`
	Env        []hclNamedBlock `hcl:"env,block"`
}

type hclActionBrandingBlock struct {
	Icon  hcl.Expression `hcl:"icon,optional"`
	Color hcl.Expression `hcl:"color,optional"`
}

type hclActionBlock struct {
	ID          string                   `hcl:"id,label"`
	Filename    hcl.Expression           `hcl:"filename,optional"`
	Name        hcl.Expression           `hcl:"name,optional"`
	Description hcl.Expression           `hcl:"description,optional"`
	Author      hcl.Expression           `hcl:"author,optional"`
	Inputs      []hclActionInputBlock    `hcl:"input,block"`
	Outputs     []hclActionOutputBlock   `hcl:"output,block"`
	Runs        []hclActionRunsBlock     `hcl:"runs,block"`
	Branding    []hclActionBrandingBlock `hcl:"branding,block"`
}

type parseConfig struct {
	Variables hclparser.VariablesConfig `hcl:"variable,block"`
	Steps     step.StepListConfig       `hcl:"step,block"`
	Jobs      []hclJobBlock             `hcl:"job,block"`
	Workflows []hclWorkflowBlock        `hcl:"workflow,block"`
	Actions   []hclActionBlock          `hcl:"action,block"`
}
