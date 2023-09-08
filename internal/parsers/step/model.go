package step

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

type StepConfig struct {
	Ref              string                  `hcl:",label"`
	Id               *IdConfig               `hcl:"id,attr"`
	If               *IfConfig               `hcl:"if,attr"`
	Name             *NameConfig             `hcl:"name,attr"`
	Uses             *UsesConfig             `hcl:"uses,attr"`
	Run              *RunConfig              `hcl:"run,attr"`
	WorkingDirectory *WorkingDirectoryConfig `hcl:"working_directory,attr"`
	Shell            *ShellConfig            `hcl:"shell,attr"`
	With             WithsConfig             `hcl:"shell,block"`
	Env              EnvsConfig              `hcl:"env,block"`
	ContinueOnError  *ContinueOnErrorConfig  `hcl:"continue_on_error,attr"`
	TimeoutMinutes   *TimeoutMinutesConfig   `hcl:"timeout_minutes,attr"`
}

type Step struct {
	Ref              string
	Id               string
	If               string
	Name             string
	Uses             string
	Run              string
	WorkingDirectory string
	Shell            string
	With             map[string]any
	Env              map[string]any
	ContinueOnError  bool
	TimeoutMinutes   uint16
}

type StepYaml struct {
	Ref              string         `yaml:"-"`
	Id               string         `yaml:"id,omitempty"`
	If               string         `yaml:"if,omitempty"`
	Name             string         `yaml:"name,omitempty"`
	Uses             string         `yaml:"uses,omitempty"`
	Run              string         `yaml:"run,omitempty"`
	WorkingDirectory string         `yaml:"working-directory,omitempty"`
	Shell            string         `yaml:"shell,omitempty"`
	With             map[string]any `yaml:"with,omitempty"`
	Env              map[string]any `yaml:"env,omitempty"`
	ContinueOnError  bool           `yaml:"continue-on-error,omitempty"`
	TimeoutMinutes   uint16         `yaml:"timeout-minutes,omitempty"`
}

func (content StepYaml) Convert() ([]byte, error) {
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
