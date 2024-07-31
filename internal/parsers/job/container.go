// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type CredentialsConfig struct {
	Username string `hcl:"username,attr"`
	Password string `hcl:"password,attr"`
}

type VariableConfig struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type EnvConfig struct {
	Variable []VariableConfig `hcl:"variable,block"`
}

type ContainerConfig struct {
	Image       string            `hcl:"image,attr"`
	Credentials CredentialsConfig `hcl:"credentials,block"`
	Env         EnvConfig         `hcl:"env,block"`
	Ports       []int16           `hcl:"ports,attr"`
	Volumes     []string          `hcl:"volumes,attr"`
	Options     string            `hcl:"options,attr"`
}

type Credentials struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Env map[string]any

type Container struct {
	Image       string      `yaml:"image,omitempty"`
	Credentials Credentials `yaml:"credentials,omitempty"`
	Env         Env         `yaml:"env,omitempty"`
	Ports       []int16     `yaml:"ports,omitempty"`
	Volumes     []string    `yaml:"volumes,omitempty"`
	Options     string      `yaml:"options,omitempty"`
}

func (config *ContainerConfig) Parse() (Container, error) {
	container := Container{}

	if config.Image != "" {
		container.Image = config.Image
	}

	if config.Credentials != (CredentialsConfig{}) {
		container.Credentials = Credentials{
			Username: config.Credentials.Username,
			Password: config.Credentials.Password,
		}
	}

	if config.Env.Variable != nil {
		envs := make(Env)

		for _, env := range config.Env.Variable {
			switch env.Value.Type().FriendlyName() {
			case "string":
				var val string
				err := gocty.FromCtyValue(env.Value, &val)
				if err != nil {
					return Container{}, err
				}
				envs[env.Name] = val
			case "number":
				var val int32
				err := gocty.FromCtyValue(env.Value, &val)
				if err != nil {
					var val float32
					err := gocty.FromCtyValue(env.Value, &val)
					if err != nil {
						return Container{}, err
					}
				}
				envs[env.Name] = val
			case "bool":
				var val bool
				err := gocty.FromCtyValue(env.Value, &val)
				if err != nil {
					return Container{}, err
				}
				envs[env.Name] = val
			}
		}

		container.Env = envs
	}

	if config.Ports != nil {
		ports := []int16{}

		container.Ports = append(ports, config.Ports...)
	}

	if config.Volumes != nil {
		volumes := []string{}

		container.Volumes = append(volumes, config.Volumes...)
	}

	if config.Options != "" {
		container.Options = config.Options
	}

	return container, nil
}
