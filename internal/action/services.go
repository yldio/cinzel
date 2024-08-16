// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"github.com/zclconf/go-cty/cty"
)

type Service struct {
	Name        string         `yaml:"-"`
	Image       string         `yaml:"image"`
	Credentials Credentials    `yaml:"credentials,omitempty"`
	Env         map[string]any `yaml:"env,omitempty"`
	Ports       []string       `yaml:"ports,omitempty"`
	Volumes     []string       `yaml:"volumes,omitempty"`
	Options     string         `yaml:"options,omitempty"`
}

type Services map[string]Service

type ServiceConfig struct {
	Name        string             `hcl:"name,label"`
	Image       *string            `hcl:"image,attr"`
	Credentials *CredentialsConfig `hcl:"credentials,block"`
	Env         *EnvConfig         `hcl:"env,block"`
	Ports       *[]string          `hcl:"ports,attr"`
	Volumes     *[]string          `hcl:"volumes,attr"`
	Options     *string            `hcl:"options,attr"`
}

type ServicesConfig []*ServiceConfig

func (config *ServicesConfig) Parse() (Services, error) {
	if config == nil {
		return nil, nil
	}

	services := Services{}

	for _, service := range *config {

		services[service.Name] = Service{
			Name:  service.Name,
			Image: *service.Image,
		}

		svr := services[service.Name]

		if service.Credentials != nil {
			credentials := Credentials{
				Username: service.Credentials.Username,
				Password: service.Credentials.Password,
			}

			svr.Credentials = credentials
		}

		if service.Env != nil && len(service.Env.Variable) != 0 {
			envs := make(map[string]any)

			for _, variable := range service.Env.Variable {
				val, err := ParseCtyValue(variable.Value, []string{
					cty.String.FriendlyName(),
					cty.Number.FriendlyName(),
					cty.Bool.FriendlyName(),
				})
				if err != nil {
					return nil, err
				}
				envs[variable.Name] = val
			}

			svr.Env = envs
		}

		if service.Ports != nil && len(*service.Ports) != 0 {
			ports := []string{}

			ports = append(ports, *service.Ports...)

			svr.Ports = ports
		}

		if service.Volumes != nil && len(*service.Volumes) != 0 {
			volumes := []string{}

			volumes = append(volumes, *service.Volumes...)

			svr.Volumes = volumes
		}

		if service.Options != nil {
			svr.Options = *service.Options
		}

		services[service.Name] = svr
	}

	return services, nil
}

func (services *Services) IsNill() bool {
	isNill := true

	for _, service := range *services {
		if service.Credentials != (Credentials{}) {
			isNill = false
		}
	}

	return isNill
}
