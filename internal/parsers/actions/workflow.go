package actions

import (
	"github.com/hashicorp/hcl/v2"
)

type WorkflowConfig struct {
	Id          string              `hcl:",label"`
	Name        *string             `hcl:"name,attr"`
	RunName     *string             `hcl:"run_name,attr"`
	On          *string             `hcl:"on,attr"`
	OnAsList    *[]string           `hcl:"on_as_list,attr"`
	OnByFilter  []*OnByFilterConfig `hcl:"on_by_filter,block"`
	Jobs        hcl.Expression      `hcl:"jobs,attr"`
	Permissions *PermissionsConfig  `hcl:"permissions,block"`
	Envs        []*EnvsConfig       `hcl:"env,block"`
	Defaults    *DefaultsConfig     `hcl:"defaults,block"`
	Concurrency *ConcurrencyConfig  `hcl:"concurrency,block"`
}

type Workflow struct {
	Id          string
	Name        string
	RunName     string
	On          any
	Jobs        []Job
	Permissions Permissions
	Envs        Envs
	Defaults    Defaults
	Concurrency Concurrency
}

type WorkflowYaml struct {
	Name        string          `yaml:"name,omitempty"`
	RunName     string          `yaml:"run-name,omitempty"`
	On          any             `yaml:"on,omitempty"`
	Permissions PermissionsYaml `yaml:"permissions,omitempty"`
	Envs        EnvsYaml        `yaml:"env,omitempty"`
	Defaults    DefaultsYaml    `yaml:"defaults,omitempty"`
	Concurrency ConcurrencyYaml `yaml:"concurrency,omitempty"`
}

func (workflow *Workflow) ConvertToYaml() (WorkflowYaml, error) {
	yaml := WorkflowYaml{
		Name:    workflow.Name,
		RunName: workflow.RunName,
		On:      workflow.On,
	}

	permissions, err := workflow.Permissions.ConvertToYaml()
	if err != nil {
		return WorkflowYaml{}, err
	}

	yaml.Permissions = permissions

	envs, err := workflow.Envs.ConvertToYaml()
	if err != nil {
		return WorkflowYaml{}, err
	}

	yaml.Envs = envs

	defaults, err := workflow.Defaults.ConvertToYaml()
	if err != nil {
		return WorkflowYaml{}, err
	}

	yaml.Defaults = defaults

	concurrency, err := workflow.Concurrency.ConvertToYaml()
	if err != nil {
		return WorkflowYaml{}, err
	}

	yaml.Concurrency = concurrency

	return yaml, nil
}
