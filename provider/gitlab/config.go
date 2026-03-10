// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"github.com/hashicorp/hcl/v2"
)

type hclRuleBlock struct {
	If           hcl.Expression `hcl:"if,optional"`
	When         hcl.Expression `hcl:"when,optional"`
	AllowFailure hcl.Expression `hcl:"allow_failure,optional"`
	Changes      hcl.Expression `hcl:"changes,optional"`
	Exists       hcl.Expression `hcl:"exists,optional"`
}

type hclReportsBlock struct {
	Body hcl.Body `hcl:",remain"`
}

type hclArtifactsBlock struct {
	Paths     hcl.Expression    `hcl:"paths,optional"`
	Exclude   hcl.Expression    `hcl:"exclude,optional"`
	ExpireIn  hcl.Expression    `hcl:"expire_in,optional"`
	Name      hcl.Expression    `hcl:"name,optional"`
	Untracked hcl.Expression    `hcl:"untracked,optional"`
	When      hcl.Expression    `hcl:"when,optional"`
	Reports   []hclReportsBlock `hcl:"reports,block"`
}

type hclCacheBlock struct {
	Key          hcl.Expression `hcl:"key,optional"`
	Paths        hcl.Expression `hcl:"paths,optional"`
	Untracked    hcl.Expression `hcl:"untracked,optional"`
	When         hcl.Expression `hcl:"when,optional"`
	Policy       hcl.Expression `hcl:"policy,optional"`
	FallbackKeys hcl.Expression `hcl:"fallback_keys,optional"`
}

type hclServiceBlock struct {
	Name       hcl.Expression `hcl:"name,optional"`
	Alias      hcl.Expression `hcl:"alias,optional"`
	Entrypoint hcl.Expression `hcl:"entrypoint,optional"`
	Command    hcl.Expression `hcl:"command,optional"`
	PullPolicy hcl.Expression `hcl:"pull_policy,optional"`
	Variables  hcl.Expression `hcl:"variables,optional"`
}

type hclVariableBlock struct {
	ID          string         `hcl:"id,label"`
	Name        hcl.Expression `hcl:"name,optional"`
	Value       hcl.Expression `hcl:"value,optional"`
	Description hcl.Expression `hcl:"description,optional"`
}

type hclJobBlock struct {
	ID            string              `hcl:"id,label"`
	Stage         hcl.Expression      `hcl:"stage,optional"`
	Image         hcl.Expression      `hcl:"image,optional"`
	Script        hcl.Expression      `hcl:"script,optional"`
	BeforeScript  hcl.Expression      `hcl:"before_script,optional"`
	AfterScript   hcl.Expression      `hcl:"after_script,optional"`
	Tags          hcl.Expression      `hcl:"tags,optional"`
	DependsOn     hcl.Expression      `hcl:"depends_on,optional"`
	Extends       hcl.Expression      `hcl:"extends,optional"`
	When          hcl.Expression      `hcl:"when,optional"`
	AllowFailure  hcl.Expression      `hcl:"allow_failure,optional"`
	Interruptible hcl.Expression      `hcl:"interruptible,optional"`
	Retry         hcl.Expression      `hcl:"retry,optional"`
	Timeout       hcl.Expression      `hcl:"timeout,optional"`
	Variables     hcl.Expression      `hcl:"variables,optional"`
	Environment   hcl.Expression      `hcl:"environment,optional"`
	Release       hcl.Expression      `hcl:"release,optional"`
	Trigger       hcl.Expression      `hcl:"trigger,optional"`
	Parallel      hcl.Expression      `hcl:"parallel,optional"`
	Coverage      hcl.Expression      `hcl:"coverage,optional"`
	ResourceGroup hcl.Expression      `hcl:"resource_group,optional"`
	Rules         []hclRuleBlock      `hcl:"rule,block"`
	Artifacts     []hclArtifactsBlock `hcl:"artifacts,block"`
	Cache         []hclCacheBlock     `hcl:"cache,block"`
	Services      []hclServiceBlock   `hcl:"service,block"`
}

type hclWorkflowBlock struct {
	Name  hcl.Expression `hcl:"name,optional"`
	Rules []hclRuleBlock `hcl:"rule,block"`
}

type hclTemplateBlock struct {
	ID            string              `hcl:"id,label"`
	Stage         hcl.Expression      `hcl:"stage,optional"`
	Image         hcl.Expression      `hcl:"image,optional"`
	Script        hcl.Expression      `hcl:"script,optional"`
	BeforeScript  hcl.Expression      `hcl:"before_script,optional"`
	AfterScript   hcl.Expression      `hcl:"after_script,optional"`
	Tags          hcl.Expression      `hcl:"tags,optional"`
	DependsOn     hcl.Expression      `hcl:"depends_on,optional"`
	Extends       hcl.Expression      `hcl:"extends,optional"`
	When          hcl.Expression      `hcl:"when,optional"`
	AllowFailure  hcl.Expression      `hcl:"allow_failure,optional"`
	Interruptible hcl.Expression      `hcl:"interruptible,optional"`
	Retry         hcl.Expression      `hcl:"retry,optional"`
	Timeout       hcl.Expression      `hcl:"timeout,optional"`
	Variables     hcl.Expression      `hcl:"variables,optional"`
	Environment   hcl.Expression      `hcl:"environment,optional"`
	Release       hcl.Expression      `hcl:"release,optional"`
	Trigger       hcl.Expression      `hcl:"trigger,optional"`
	Parallel      hcl.Expression      `hcl:"parallel,optional"`
	Coverage      hcl.Expression      `hcl:"coverage,optional"`
	ResourceGroup hcl.Expression      `hcl:"resource_group,optional"`
	Rules         []hclRuleBlock      `hcl:"rule,block"`
	Artifacts     []hclArtifactsBlock `hcl:"artifacts,block"`
	Cache         []hclCacheBlock     `hcl:"cache,block"`
	Services      []hclServiceBlock   `hcl:"service,block"`
}

type hclDefaultBlock struct {
	Image         hcl.Expression    `hcl:"image,optional"`
	BeforeScript  hcl.Expression    `hcl:"before_script,optional"`
	AfterScript   hcl.Expression    `hcl:"after_script,optional"`
	Tags          hcl.Expression    `hcl:"tags,optional"`
	Interruptible hcl.Expression    `hcl:"interruptible,optional"`
	Retry         hcl.Expression    `hcl:"retry,optional"`
	Timeout       hcl.Expression    `hcl:"timeout,optional"`
	Cache         []hclCacheBlock   `hcl:"cache,block"`
	Services      []hclServiceBlock `hcl:"service,block"`
}

type hclIncludeBlock struct {
	Local     hcl.Expression `hcl:"local,optional"`
	Project   hcl.Expression `hcl:"project,optional"`
	File      hcl.Expression `hcl:"file,optional"`
	Ref       hcl.Expression `hcl:"ref,optional"`
	Remote    hcl.Expression `hcl:"remote,optional"`
	Template  hcl.Expression `hcl:"template,optional"`
	Component hcl.Expression `hcl:"component,optional"`
	Inputs    hcl.Expression `hcl:"inputs,optional"`
}

type parseConfig struct {
	Stages    []string           `hcl:"stages,optional"`
	Variables []hclVariableBlock `hcl:"variable,block"`
	Jobs      []hclJobBlock      `hcl:"job,block"`
	Workflow  []hclWorkflowBlock `hcl:"workflow,block"`
	Templates []hclTemplateBlock `hcl:"template,block"`
	Includes  []hclIncludeBlock  `hcl:"include,block"`
	Default   []hclDefaultBlock  `hcl:"default,block"`
}
