// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type Container struct {
	Image       string      `yaml:"image,omitempty"`
	Credentials Credentials `yaml:"credentials,omitempty"`
	Env         Env         `yaml:"env,omitempty"`
	Ports       []int16     `yaml:"ports,omitempty"`
	Volumes     []string    `yaml:"volumes,omitempty"`
	Options     string      `yaml:"options,omitempty"`
}

type ContainerConfig struct {
	Image       string            `hcl:"image,attr"`
	Credentials CredentialsConfig `hcl:"credentials,block"`
	Env         EnvConfig         `hcl:"env,block"`
	Ports       []int16           `hcl:"ports,attr"`
	Volumes     []string          `hcl:"volumes,attr"`
	Options     string            `hcl:"options,attr"`
}

func (config *ContainerConfig) Parse() (Container, error) {
	if config == nil {
		return Container{}, nil
	}

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
		envs, err := config.Env.Parse()
		if err != nil {
			return Container{}, nil
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

func (container *Container) IsNill() bool {
	isNill := true

	if container.Credentials != (Credentials{}) {
		isNill = false
	}

	if container.Env != nil {
		isNill = false
	}

	return isNill
}
