package job

import (
	"bytes"

	"github.com/goccy/go-yaml"
)

type JobConfig struct {
	Ref            string                `hcl:",label"`
	TimeoutMinutes *TimeoutMinutesConfig `hcl:"timeout_minutes,attr"`
	Strategy       *StrategyConfig       `hcl:"strategy,block"`
	Container      *ContainerConfig      `hcl:"container,block"`
	Services       *ServicesConfig       `hcl:"services,block"`
}

type StepYaml struct {
	Ref            string `yaml:"-"`
	TimeoutMinutes uint16 `yaml:"timeout-minutes,omitempty"`
	Strategy       any    `yaml:"strategy,omitempty"`
	Container      any    `yaml:"container,omitempty"`
	Services       any    `yaml:"services,omitempty"`
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
