package workflow

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

type WorkflowConfig struct {
	Id          string              `hcl:",label"`
	Name        *NameConfig         `hcl:"name,attr"`
	RunName     *RunNameConfig      `hcl:"run_name,attr"`
	On          *OnConfig           `hcl:"on,attr"`
	OnAsList    *OnAsListConfig     `hcl:"on_as_list,attr"`
	OnByFilter  []*OnByFilterConfig `hcl:"on_by_filter,block"`
	Permissions *PermissionsConfig  `hcl:"permissions,block"`
	Env         EnvsConfig          `hcl:"env,block"`
	Concurrency *ConcurrencyConfig  `hcl:"concurrency,block"`
}

type Workflow struct {
	Id      string
	Name    string
	RunName string
	On      any
}

type WorkflowYaml struct {
	Name    string `yaml:"name,omitempty"`
	RunName string `yaml:"run-name,omitempty"`
	On      any    `yaml:"on"`
}

func (content WorkflowYaml) Convert() ([]byte, error) {
	return Convert(content)
}

func Convert(content any) ([]byte, error) {
	out, err := yaml.Marshal(content)
	if err != nil {
		return []byte{}, err
	}

	// Please link to https://github.com/go-yaml/yaml?tab=readme-ov-file#yaml-support-for-the-go-language
	// `atos` uses `any` so we need this "hack" to clean `"on":` to just `on:`.
	filteredOut := bytes.Replace(out, []byte("\"on\""), []byte("on"), -1)

	return filteredOut, nil
}

func ConvertFromHcl() {}
