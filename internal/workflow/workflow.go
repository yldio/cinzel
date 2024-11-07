// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package workflow

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/action"
	"github.com/yldio/acto/internal/actoerrors"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/job"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type Workflows []*Workflow

type Workflow struct {
	Id          string              `yaml:"-"`
	Filename    string              `yaml:"-" hcl:"filename"`
	Name        string              `yaml:"name,omitempty" hcl:"name"`
	RunName     string              `yaml:"run-name,omitempty" hcl:"run_name"`
	On          *action.On          `yaml:"on" hcl:"on"`
	Permissions *action.Permissions `yaml:"permissions,omitempty" hcl:"permissions"`
	Env         *action.Envs        `yaml:"env,omitempty" hcl:"env"`
	Defaults    *action.Defaults    `yaml:"defaults,omitempty" hcl:"defaults"`
	Concurrency *action.Concurrency `yaml:"concurrency,omitempty" hcl:"concurrency"`
	Jobs        job.Jobs            `yaml:"jobs" hcl:"job"`
	JobsIds     []string            `yaml:"-"`
}

type WorkflowsConfig []WorkflowConfig

type WorkflowConfig struct {
	Id          string                    `hcl:"id,label"`
	Filename    hcl.Expression            `hcl:"filename,attr"`
	Name        hcl.Expression            `hcl:"name,attr"`
	RunName     hcl.Expression            `hcl:"run_name,attr"`
	On          action.EventsConfig       `hcl:"on,block"`
	Permissions *action.PermissionsConfig `hcl:"permissions,block"`
	Env         action.EnvsConfig         `hcl:"env,block"`
	Defaults    *action.DefaultsConfig    `hcl:"defaults,block"`
	Concurrency *action.ConcurrencyConfig `hcl:"concurrency,block"`
	Jobs        hcl.Expression            `hcl:"jobs,attr"`
}

func (config *WorkflowConfig) unwrapFilename(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'filename' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapFilename(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'filename' must be a string")
	}
}

func (config *WorkflowConfig) parseFilename() (*string, error) {
	acto := actoparser.NewActo(config.Filename)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapFilename(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *WorkflowConfig) unwrapName(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'name' must be a string")
	}
}

func (config *WorkflowConfig) parseName() (*string, error) {
	acto := actoparser.NewActo(config.Name)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapName(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *WorkflowConfig) unwrapRunName(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapRunName(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'run_name' must be a string")
	}
}

func (config *WorkflowConfig) parseRunName() (*string, error) {
	acto := actoparser.NewActo(config.RunName)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapName(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *WorkflowConfig) parseOn() (*action.On, error) {
	on, err := config.On.Parse()
	if err != nil {
		return nil, err
	}

	return on, nil
}

func (config *WorkflowConfig) parsePermissions() (*action.Permissions, error) {
	permissions, err := config.Permissions.Parse()
	if err != nil {
		return nil, err
	}

	return permissions, nil
}

func (config *WorkflowConfig) parseEnvs() (*action.Envs, error) {
	env, err := config.Env.Parse()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (config *WorkflowConfig) parseDefaults() (*action.Defaults, error) {
	defaults, err := config.Defaults.Parse()
	if err != nil {
		return nil, err
	}

	return defaults, nil
}

func (config *WorkflowConfig) parseConcurrency() (*action.Concurrency, error) {
	concurrency, err := config.Concurrency.Parse()
	if err != nil {
		return nil, err
	}

	if concurrency == nil {
		return nil, nil
	}

	return concurrency, nil
}

func (config *WorkflowConfig) unwrapJobsIds(acto *actoparser.Acto) (*[]string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'jobs' must be a list of jobs relation")
	case []actoparser.ActoVariableRef:
		list := []string{}
		for _, jobRef := range resultValue {
			if jobRef.Name != "job" {
				return nil, errors.New("invalid job reference, should be job.<job-identifier>")
			}

			list = append(list, jobRef.Attr)
		}

		return &list, nil
	default:
		return nil, errors.New("attribute 'jobs' must be a list of jobs relation")
	}
}

func (config *WorkflowConfig) parseJobsIds() (*[]string, error) {
	acto := actoparser.NewActo(config.Jobs)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	jobsIds, err := config.unwrapJobsIds(acto)
	if err != nil {
		return nil, err
	}

	if len(*jobsIds) == 0 {
		return nil, errors.New("attribute 'jobs' cannot be empty")
	}

	return jobsIds, nil
}

func (config *WorkflowConfig) Parse() (*Workflow, error) {
	if config == nil {
		return nil, nil
	}

	if config.Id == "" {
		return nil, errors.New("a workflow needs to have an id")
	}

	workflow := Workflow{
		Id: config.Id,
	}

	filename, err := config.parseFilename()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	workflow.Filename = *filename

	name, err := config.parseName()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	if name != nil {
		workflow.Name = *name
	}

	runName, err := config.parseRunName()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	if runName != nil {
		workflow.RunName = *runName
	}

	on, err := config.parseOn()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	workflow.On = on

	permissions, err := config.parsePermissions()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	if permissions != nil {
		workflow.Permissions = permissions
	}

	envs, err := config.parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	if envs != nil {
		workflow.Env = envs
	}

	defaults, err := config.parseDefaults()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	if defaults != nil {
		workflow.Defaults = defaults
	}

	concurrency, err := config.parseConcurrency()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	if concurrency != nil {
		workflow.Concurrency = concurrency
	}

	jobsIds, err := config.parseJobsIds()
	if err != nil {
		return nil, fmt.Errorf("error in workflow '%s': %w, %w", workflow.Id, err, actoerrors.ErrOpenIssue)
	}

	workflow.JobsIds = *jobsIds

	return &workflow, nil
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

func (workflow *Workflow) Decode() ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()

	label := actoparser.ToSnakeCase(workflow.Id)

	workflowBlock := rootBody.AppendNewBlock("workflow", []string{label})

	workflowBody := workflowBlock.Body()

	workflowBody.SetAttributeValue("filename", cty.StringVal(label))

	if workflow.Name != "" {
		nameAttr, err := actoparser.GetHclTag(*workflow, "Name")
		if err != nil {
			return []byte(``), err
		}

		if len(workflowBody.Blocks()) > 0 || len(workflowBody.Attributes()) > 0 {
			workflowBody.AppendNewline()
		}

		workflowBody.SetAttributeValue(nameAttr, cty.StringVal(workflow.Name))
	}

	if workflow.RunName != "" {
		runNameAttr, err := actoparser.GetHclTag(*workflow, "RunName")
		if err != nil {
			return []byte(``), err
		}

		if len(workflowBody.Blocks()) > 0 || len(workflowBody.Attributes()) > 0 {
			workflowBody.AppendNewline()
		}

		workflowBody.SetAttributeValue(runNameAttr, cty.StringVal(workflow.RunName))
	}

	if workflow.On != nil {
		onAttr, err := actoparser.GetHclTag(*workflow, "On")
		if err != nil {
			return []byte(``), err
		}

		if err := workflow.On.Decode(workflowBody, onAttr); err != nil {
			return []byte(``), err
		}
	}

	if workflow.Permissions != nil {
		permissionsAttr, err := actoparser.GetHclTag(*workflow, "Permissions")
		if err != nil {
			return []byte(``), err
		}

		if err := workflow.Permissions.Decode(workflowBody, permissionsAttr); err != nil {
			return []byte(``), err
		}
	}

	if workflow.Env != nil {
		for name, env := range *workflow.Env {

			if len(workflowBody.Blocks()) > 0 {
				workflowBody.AppendNewline()
			}

			envBlock := workflowBody.AppendNewBlock("env", nil)

			envBody := envBlock.Body()
			envBody.SetAttributeValue("name", cty.StringVal(name))

			switch e := env.(type) {
			case string:
				envBody.SetAttributeValue("value", cty.StringVal(e))
			}
		}
	}

	if workflow.Defaults != nil {
		if len(workflowBody.Blocks()) > 0 {
			workflowBody.AppendNewline()
		}

		attr, err := actoparser.GetHclTag(*workflow, "Defaults")
		if err != nil {
			return []byte(``), err
		}

		defaultsBlock := workflowBody.AppendNewBlock(attr, nil)
		defaultsBody := defaultsBlock.Body()
		attr, err = actoparser.GetHclTag(*workflow.Defaults, "Run")
		if err != nil {
			return []byte(``), err
		}

		runBlock := defaultsBody.AppendNewBlock(attr, nil)
		runBody := runBlock.Body()

		if workflow.Defaults.Run.Shell != nil {
			attr, err := actoparser.GetHclTag(*workflow.Defaults.Run, "Shell")
			if err != nil {
				return []byte(``), err
			}

			runBody.SetAttributeValue(attr, cty.StringVal(*workflow.Defaults.Run.Shell))
		}

		if workflow.Defaults.Run.WorkingDirectory != nil {
			attr, err := actoparser.GetHclTag(*workflow.Defaults.Run, "WorkingDirectory")
			if err != nil {
				return []byte(``), err
			}

			runBody.SetAttributeValue(attr, cty.StringVal(*workflow.Defaults.Run.WorkingDirectory))
		}
	}

	if workflow.Concurrency != nil {
		attr, err := actoparser.GetHclTag(*workflow, "Concurrency")
		if err != nil {
			return []byte(``), err
		}

		if len(workflowBody.Blocks()) > 0 {
			workflowBody.AppendNewline()
		}

		concurrencyBlock := workflowBody.AppendNewBlock(attr, nil)
		concurrencyBody := concurrencyBlock.Body()

		attr, err = actoparser.GetHclTag(*workflow.Concurrency, "Group")
		if err != nil {
			return []byte(``), err
		}

		if workflow.Concurrency.Group != nil {
			attr, err := actoparser.GetHclTag(*workflow.Concurrency, "Group")
			if err != nil {
				return []byte(``), err
			}

			concurrencyBody.SetAttributeValue(attr, cty.StringVal(*workflow.Concurrency.Group))
		}

		if workflow.Concurrency.CancelInProgress != nil {
			attr, err := actoparser.GetHclTag(*workflow.Concurrency, "CancelInProgress")
			if err != nil {
				return []byte(``), err
			}

			concurrencyBody.SetAttributeValue(attr, cty.BoolVal(*workflow.Concurrency.CancelInProgress))
		}
	}

	if workflow.Jobs != nil {
		jobAttr, err := actoparser.GetHclTag(*workflow, "Jobs")
		if err != nil {
			return []byte(``), err
		}

		for id, j := range workflow.Jobs {

			j.Id = fmt.Sprintf("%s-%s", workflow.Id, id)

			if err := j.Decode(rootBody, jobAttr); err != nil {
				return []byte(``), err
			}
		}
	}

	return f.Bytes(), nil
}
