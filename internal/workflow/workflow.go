// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package workflow

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/job"
)

type Workflows []Workflow

type Workflow struct {
	Id          string              `yaml:"-"`
	Filename    string              `yaml:"-"`
	Name        *string             `yaml:"name,omitempty"`
	RunName     *string             `yaml:"run-name,omitempty"`
	On          action.On           `yaml:"on"`
	Permissions *action.Permissions `yaml:"permissions,omitempty"`
	Env         *action.Env         `yaml:"env,omitempty"`
	Defaults    *action.Defaults    `yaml:"defaults,omitempty"`
	Concurrency *action.Concurrency `yaml:"concurrency,omitempty"`
	Jobs        map[string]job.Job  `yaml:"jobs"`
	JobsIds     []string            `yaml:"-"`
}

type WorkflowsConfig []WorkflowConfig

type WorkflowConfig struct {
	Id          string                    `hcl:"id,label"`
	Filename    *string                   `hcl:"filename,attr"`
	Name        *action.NameConfig        `hcl:"name,attr"`
	RunName     *action.RunNameConfig     `hcl:"run_name,attr"`
	On          action.OnsConfig          `hcl:"on,block"`
	Permissions *action.PermissionsConfig `hcl:"permissions,block"`
	Env         *action.EnvConfig         `hcl:"env,block"`
	Defaults    *action.DefaultsConfig    `hcl:"defaults,block"`
	Concurrency *action.ConcurrencyConfig `hcl:"concurrency,block"`
	Jobs        hcl.Expression            `hcl:"jobs,attr"`
}

func (config *WorkflowConfig) parseId(workflow *Workflow) error {
	if config.Id == "" {
		return nil
	}

	workflow.Id = config.Id

	return nil
}

func (config *WorkflowConfig) parseName(workflow *Workflow) error {
	name, err := config.Name.Parse()
	if err != nil {
		return err
	}

	if name == "" {
		return nil
	}

	workflow.Name = &name

	return nil
}

func (config *WorkflowConfig) parseRunName(workflow *Workflow) error {
	runName, err := config.RunName.Parse()
	if err != nil {
		return err
	}

	if runName == "" {
		return nil
	}

	workflow.RunName = &runName

	return nil
}

func (config *WorkflowConfig) parseOn(workflow *Workflow) error {
	on, err := config.On.Parse(workflow.Id)
	if err != nil {
		return err
	}

	workflow.On = on

	return nil
}

func (config *WorkflowConfig) parsePermissions(workflow *Workflow) error {
	permissions, err := config.Permissions.Parse()
	if err != nil {
		return err
	}

	if permissions == (action.Permissions{}) {
		return nil
	}

	workflow.Permissions = &permissions

	return nil
}

func (config *WorkflowConfig) parseEnv(workflow *Workflow) error {
	env, err := config.Env.Parse()
	if err != nil {
		return err
	}

	if env == nil {
		return nil
	}

	workflow.Env = &env

	return nil
}

func (config *WorkflowConfig) parseDefaults(workflow *Workflow) error {
	defaults, err := config.Defaults.Parse()
	if err != nil {
		return err
	}

	if defaults == (action.Defaults{}) {
		return nil
	}

	workflow.Defaults = &defaults

	return nil
}

func (config *WorkflowConfig) parseConcurrency(workflow *Workflow) error {
	concurrency, err := config.Concurrency.Parse()
	if err != nil {
		return err
	}

	if concurrency == (action.Concurrency{}) {
		return nil
	}

	workflow.Concurrency = &concurrency

	return nil
}

func (config *WorkflowConfig) parseJobs(workflow *Workflow) error {
	if config.Jobs == nil {
		return nil
	}

	val, diags := config.Jobs.Value(nil)
	if diags.HasErrors() {
		// TODO: understand the reason why it has diags
		// return nil, errors.New(diags[0].Detail)
	}
	if val.IsNull() {
		return actoerrors.ErrWorkflowEmptyJobs(workflow.Id)
	}

	exprs, diags := hcl.ExprList(config.Jobs)
	if diags.HasErrors() {
		return actoerrors.ProcessHCLDiags(diags)
	}

	jobsIds := []string{}

	for _, expr := range exprs {
		traversal, diags := hcl.AbsTraversalForExpr(expr)
		if diags.HasErrors() {
			return actoerrors.ProcessHCLDiags(diags)
		}

		for _, traverser := range traversal {
			switch tJob := traverser.(type) {
			case hcl.TraverseRoot:
				if tJob.Name != "job" {
					return errors.New("jobs require a job relationship only")
				}
			case hcl.TraverseAttr:
				jobsIds = append(jobsIds, tJob.Name)
			}
		}
	}

	if len(jobsIds) == 0 {
		return errors.New("requires at least one job")
	}

	workflow.JobsIds = jobsIds

	return nil
}

func (config *WorkflowConfig) Parse() (Workflow, error) {
	if config == nil {
		return Workflow{}, nil
	}

	if config.Filename == nil {
		return Workflow{}, actoerrors.ErrWorkflowFilenameRequired
	}

	workflow := Workflow{
		Filename: *config.Filename,
	}

	if err := config.parseId(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseName(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseRunName(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseOn(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parsePermissions(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseEnv(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseDefaults(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseConcurrency(&workflow); err != nil {
		return Workflow{}, err
	}

	if err := config.parseJobs(&workflow); err != nil {
		return Workflow{}, err
	}

	return workflow, nil
}

func (config *WorkflowsConfig) Parse() (Workflows, error) {
	workflows := Workflows{}

	for _, workflow := range *config {
		parsedWorkflow, err := workflow.Parse()
		if err != nil {
			return Workflows{}, err
		}

		workflows = append(workflows, parsedWorkflow)
	}

	return workflows, nil
}

func (workflows *Workflows) PostParse(parsedJobs job.Jobs) error {
	for idx, workflow := range *workflows {
		for _, jobId := range workflow.JobsIds {
			if len(workflow.JobsIds) == 0 {
				return fmt.Errorf("workflow %s requires at least one job", workflow.Id)
			}

			for _, parsedJob := range parsedJobs {
				if parsedJob.Id != jobId {
					continue
				}

				if workflow.Jobs == nil {
					workflow.Jobs = make(map[string]job.Job)
				}

				workflow.Jobs[jobId] = parsedJob
			}
		}

		if len(workflow.Jobs) == 0 && len(workflow.JobsIds) > 0 {
			return errors.New("some jobs do not exist")
		}

		(*workflows)[idx] = workflow
	}
	return nil
}
