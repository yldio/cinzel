// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"github.com/hashicorp/hcl/v2"
)

type hclRuleBlock struct {
	If            hcl.Expression `hcl:"if,optional"`
	When          hcl.Expression `hcl:"when,optional"`
	AllowFailure  hcl.Expression `hcl:"allow_failure,optional"`
	Changes       hcl.Expression `hcl:"changes,optional"`
	Exists        hcl.Expression `hcl:"exists,optional"`
	Variables     hcl.Expression `hcl:"variables,optional"`
	Needs         hcl.Expression `hcl:"needs,optional"`
	StartIn       hcl.Expression `hcl:"start_in,optional"`
	Interruptible hcl.Expression `hcl:"interruptible,optional"`
	AutoCancel    hcl.Expression `hcl:"auto_cancel,optional"`
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
	ExposeAs  hcl.Expression    `hcl:"expose_as,optional"`
	Public    hcl.Expression    `hcl:"public,optional"`
	Access    hcl.Expression    `hcl:"access,optional"`
	Reports   []hclReportsBlock `hcl:"reports,block"`
}

// hclNeedBlock is the object form of a YAML "needs" entry. The list form,
// which names jobs only, stays the "depends_on" attribute.
type hclNeedBlock struct {
	Job       hcl.Expression `hcl:"job,optional"`
	Artifacts hcl.Expression `hcl:"artifacts,optional"`
	Optional  hcl.Expression `hcl:"optional,optional"`
	Parallel  hcl.Expression `hcl:"parallel,optional"`
	Project   hcl.Expression `hcl:"project,optional"`
	Ref       hcl.Expression `hcl:"ref,optional"`
	Pipeline  hcl.Expression `hcl:"pipeline,optional"`
}

type hclCacheBlock struct {
	Key          hcl.Expression `hcl:"key,optional"`
	Paths        hcl.Expression `hcl:"paths,optional"`
	Untracked    hcl.Expression `hcl:"untracked,optional"`
	When         hcl.Expression `hcl:"when,optional"`
	Policy       hcl.Expression `hcl:"policy,optional"`
	FallbackKeys hcl.Expression `hcl:"fallback_keys,optional"`
	Unprotect    hcl.Expression `hcl:"unprotect,optional"`
}

type hclServiceBlock struct {
	Name       hcl.Expression `hcl:"name,optional"`
	Alias      hcl.Expression `hcl:"alias,optional"`
	Entrypoint hcl.Expression `hcl:"entrypoint,optional"`
	Command    hcl.Expression `hcl:"command,optional"`
	PullPolicy hcl.Expression `hcl:"pull_policy,optional"`
	Variables  hcl.Expression `hcl:"variables,optional"`
	Docker     hcl.Expression `hcl:"docker,optional"`
	Kubernetes hcl.Expression `hcl:"kubernetes,optional"`
}

type hclVariableBlock struct {
	ID          string         `hcl:"id,label"`
	Name        hcl.Expression `hcl:"name,optional"`
	Value       hcl.Expression `hcl:"value,optional"`
	Description hcl.Expression `hcl:"description,optional"`
	Options     hcl.Expression `hcl:"options,optional"`
	Expand      hcl.Expression `hcl:"expand,optional"`
}

type hclJobBlock struct {
	ID                 string         `hcl:"id,label"`
	Key                hcl.Expression `hcl:"id,optional"`
	Stage              hcl.Expression `hcl:"stage,optional"`
	Image              hcl.Expression `hcl:"image,optional"`
	Script             hcl.Expression `hcl:"script,optional"`
	BeforeScript       hcl.Expression `hcl:"before_script,optional"`
	AfterScript        hcl.Expression `hcl:"after_script,optional"`
	Tags               hcl.Expression `hcl:"tags,optional"`
	DependsOn          hcl.Expression `hcl:"depends_on,optional"`
	Extends            hcl.Expression `hcl:"extends,optional"`
	When               hcl.Expression `hcl:"when,optional"`
	AllowFailure       hcl.Expression `hcl:"allow_failure,optional"`
	Interruptible      hcl.Expression `hcl:"interruptible,optional"`
	Retry              hcl.Expression `hcl:"retry,optional"`
	Timeout            hcl.Expression `hcl:"timeout,optional"`
	Variables          hcl.Expression `hcl:"variables,optional"`
	Environment        hcl.Expression `hcl:"environment,optional"`
	Release            hcl.Expression `hcl:"release,optional"`
	Trigger            hcl.Expression `hcl:"trigger,optional"`
	Parallel           hcl.Expression `hcl:"parallel,optional"`
	Coverage           hcl.Expression `hcl:"coverage,optional"`
	ResourceGroup      hcl.Expression `hcl:"resource_group,optional"`
	Dependencies       hcl.Expression `hcl:"dependencies,optional"`
	StartIn            hcl.Expression `hcl:"start_in,optional"`
	Identity           hcl.Expression `hcl:"identity,optional"`
	ManualConfirmation hcl.Expression `hcl:"manual_confirmation,optional"`
	Inherit            hcl.Expression `hcl:"inherit,optional"`
	Secrets            hcl.Expression `hcl:"secrets,optional"`
	IDTokens           hcl.Expression `hcl:"id_tokens,optional"`
	Hooks              hcl.Expression `hcl:"hooks,optional"`
	Pages              hcl.Expression `hcl:"pages,optional"`
	Run                hcl.Expression `hcl:"run,optional"`
	DastConfiguration  hcl.Expression `hcl:"dast_configuration,optional"`
	Inputs             hcl.Expression `hcl:"inputs,optional"`
	Publish            hcl.Expression `hcl:"publish,optional"`
	Only               hcl.Expression `hcl:"only,optional"`
	Except             hcl.Expression `hcl:"except,optional"`
	// A keyword below is a block, which has no empty spelling, so each
	// carries an attribute holding the explicit empty list GitLab reads
	// as "override whatever this job would inherit".
	EmptyRules     hcl.Expression      `hcl:"rules,optional"`
	EmptyCache     hcl.Expression      `hcl:"cache,optional"`
	EmptyServices  hcl.Expression      `hcl:"services,optional"`
	EmptyArtifacts hcl.Expression      `hcl:"artifacts,optional"`
	Needs          []hclNeedBlock      `hcl:"need,block"`
	Rules          []hclRuleBlock      `hcl:"rule,block"`
	Artifacts      []hclArtifactsBlock `hcl:"artifacts,block"`
	Cache          []hclCacheBlock     `hcl:"cache,block"`
	Services       []hclServiceBlock   `hcl:"service,block"`
}

type hclWorkflowBlock struct {
	Name       hcl.Expression `hcl:"name,optional"`
	AutoCancel hcl.Expression `hcl:"auto_cancel,optional"`
	EmptyRules hcl.Expression `hcl:"rules,optional"`
	Rules      []hclRuleBlock `hcl:"rule,block"`
}

type hclTemplateBlock struct {
	ID                 string         `hcl:"id,label"`
	Key                hcl.Expression `hcl:"id,optional"`
	Stage              hcl.Expression `hcl:"stage,optional"`
	Image              hcl.Expression `hcl:"image,optional"`
	Script             hcl.Expression `hcl:"script,optional"`
	BeforeScript       hcl.Expression `hcl:"before_script,optional"`
	AfterScript        hcl.Expression `hcl:"after_script,optional"`
	Tags               hcl.Expression `hcl:"tags,optional"`
	DependsOn          hcl.Expression `hcl:"depends_on,optional"`
	Extends            hcl.Expression `hcl:"extends,optional"`
	When               hcl.Expression `hcl:"when,optional"`
	AllowFailure       hcl.Expression `hcl:"allow_failure,optional"`
	Interruptible      hcl.Expression `hcl:"interruptible,optional"`
	Retry              hcl.Expression `hcl:"retry,optional"`
	Timeout            hcl.Expression `hcl:"timeout,optional"`
	Variables          hcl.Expression `hcl:"variables,optional"`
	Environment        hcl.Expression `hcl:"environment,optional"`
	Release            hcl.Expression `hcl:"release,optional"`
	Trigger            hcl.Expression `hcl:"trigger,optional"`
	Parallel           hcl.Expression `hcl:"parallel,optional"`
	Coverage           hcl.Expression `hcl:"coverage,optional"`
	ResourceGroup      hcl.Expression `hcl:"resource_group,optional"`
	Dependencies       hcl.Expression `hcl:"dependencies,optional"`
	StartIn            hcl.Expression `hcl:"start_in,optional"`
	Identity           hcl.Expression `hcl:"identity,optional"`
	ManualConfirmation hcl.Expression `hcl:"manual_confirmation,optional"`
	Inherit            hcl.Expression `hcl:"inherit,optional"`
	Secrets            hcl.Expression `hcl:"secrets,optional"`
	IDTokens           hcl.Expression `hcl:"id_tokens,optional"`
	Hooks              hcl.Expression `hcl:"hooks,optional"`
	Pages              hcl.Expression `hcl:"pages,optional"`
	Run                hcl.Expression `hcl:"run,optional"`
	DastConfiguration  hcl.Expression `hcl:"dast_configuration,optional"`
	Inputs             hcl.Expression `hcl:"inputs,optional"`
	Publish            hcl.Expression `hcl:"publish,optional"`
	Only               hcl.Expression `hcl:"only,optional"`
	Except             hcl.Expression `hcl:"except,optional"`
	// A keyword below is a block, which has no empty spelling, so each
	// carries an attribute holding the explicit empty list GitLab reads
	// as "override whatever this job would inherit".
	EmptyRules     hcl.Expression      `hcl:"rules,optional"`
	EmptyCache     hcl.Expression      `hcl:"cache,optional"`
	EmptyServices  hcl.Expression      `hcl:"services,optional"`
	EmptyArtifacts hcl.Expression      `hcl:"artifacts,optional"`
	Needs          []hclNeedBlock      `hcl:"need,block"`
	Rules          []hclRuleBlock      `hcl:"rule,block"`
	Artifacts      []hclArtifactsBlock `hcl:"artifacts,block"`
	Cache          []hclCacheBlock     `hcl:"cache,block"`
	Services       []hclServiceBlock   `hcl:"service,block"`
}

type hclDefaultBlock struct {
	Image          hcl.Expression      `hcl:"image,optional"`
	BeforeScript   hcl.Expression      `hcl:"before_script,optional"`
	AfterScript    hcl.Expression      `hcl:"after_script,optional"`
	Tags           hcl.Expression      `hcl:"tags,optional"`
	Interruptible  hcl.Expression      `hcl:"interruptible,optional"`
	Retry          hcl.Expression      `hcl:"retry,optional"`
	Timeout        hcl.Expression      `hcl:"timeout,optional"`
	IDTokens       hcl.Expression      `hcl:"id_tokens,optional"`
	Hooks          hcl.Expression      `hcl:"hooks,optional"`
	EmptyCache     hcl.Expression      `hcl:"cache,optional"`
	EmptyServices  hcl.Expression      `hcl:"services,optional"`
	EmptyArtifacts hcl.Expression      `hcl:"artifacts,optional"`
	Cache          []hclCacheBlock     `hcl:"cache,block"`
	Services       []hclServiceBlock   `hcl:"service,block"`
	Artifacts      []hclArtifactsBlock `hcl:"artifacts,block"`
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
	Rules     hcl.Expression `hcl:"rules,optional"`
	Integrity hcl.Expression `hcl:"integrity,optional"`
}

type parseConfig struct {
	// GitLab still accepts these five at the top level, where they mean what
	// the same keys mean under "default". They are kept where they were
	// written rather than folded into a default block, since GitLab does not
	// document which wins when a pipeline has both.
	Image        hcl.Expression `hcl:"image,optional"`
	BeforeScript hcl.Expression `hcl:"before_script,optional"`
	AfterScript  hcl.Expression `hcl:"after_script,optional"`
	Cache        hcl.Expression `hcl:"cache,optional"`
	Services     hcl.Expression `hcl:"services,optional"`

	Stages    []string           `hcl:"stages,optional"`
	Variables []hclVariableBlock `hcl:"variable,block"`
	Jobs      []hclJobBlock      `hcl:"job,block"`
	Workflow  []hclWorkflowBlock `hcl:"workflow,block"`
	Templates []hclTemplateBlock `hcl:"template,block"`
	Includes  []hclIncludeBlock  `hcl:"include,block"`
	Default   []hclDefaultBlock  `hcl:"default,block"`
	Spec      []hclSpecBlock     `hcl:"spec,block"`
}

// hclSpecBlock is the pipeline's "spec" header, which GitLab requires to sit in
// a document of its own ahead of the rest of the configuration.
type hclSpecBlock struct {
	Inputs      hcl.Expression `hcl:"inputs,optional"`
	Include     hcl.Expression `hcl:"include,optional"`
	Component   hcl.Expression `hcl:"component,optional"`
	Description hcl.Expression `hcl:"description,optional"`
}
