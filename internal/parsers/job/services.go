// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type ServiceCredentialsConfig struct {
	Username string `hcl:"username,attr"`
	Password string `hcl:"password,attr"`
}

type ServiceVariableConfig struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type ServiceEnvConfig struct {
	Variable []ServiceVariableConfig `hcl:"variable,block"`
}

type ServiceConfig struct {
	Name        string                   `hcl:"name,label"`
	Image       string                   `hcl:"image,attr"`
	Credentials ServiceCredentialsConfig `hcl:"credentials,block"`
	Env         ServiceEnvConfig         `hcl:"env,block"`
	Ports       []string                 `hcl:"ports,attr"`
	Volumes     []string                 `hcl:"volumes,attr"`
	Options     string                   `hcl:"options,attr"`
}

type ServicesConfig []ServiceConfig

type ServiceCredentials struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Service struct {
	Name        string             `yaml:"-"`
	Image       string             `yaml:"image"`
	Credentials ServiceCredentials `yaml:"credentials,omitempty"`
	Env         map[string]any     `yaml:"env,omitempty"`
	Ports       []string           `yaml:"ports,omitempty"`
	Volumes     []string           `yaml:"volumes,omitempty"`
	Options     string             `yaml:"options,omitempty"`
}

type Services map[string]Service

func (config *ServicesConfig) Parse() (Job, error) {
	services := Services{}

	for _, service := range *config {

		services[service.Name] = Service{
			Name:  service.Name,
			Image: service.Image,
		}

		svr := services[service.Name]

		if service.Credentials != (ServiceCredentialsConfig{}) {
			credentials := ServiceCredentials{
				Username: service.Credentials.Username,
				Password: service.Credentials.Password,
			}

			svr.Credentials = credentials
		}

		if len(service.Env.Variable) != 0 {
			envs := make(map[string]any)

			for _, variable := range service.Env.Variable {
				switch variable.Value.Type().FriendlyName() {
				case "string":
					var val string
					err := gocty.FromCtyValue(variable.Value, &val)
					if err != nil {
						return Job{}, err
					}
					envs[variable.Name] = val
				case "number":
					var val int32
					err := gocty.FromCtyValue(variable.Value, &val)
					if err != nil {
						var val float32
						err := gocty.FromCtyValue(variable.Value, &val)
						if err != nil {
							return Job{}, err
						}
					}
					envs[variable.Name] = val
				case "bool":
					var val bool
					err := gocty.FromCtyValue(variable.Value, &val)
					if err != nil {
						return Job{}, err
					}
					envs[variable.Name] = val
				}
			}

			svr.Env = envs
		}

		if len(service.Ports) != 0 {
			ports := []string{}
			ports = append(ports, service.Ports...)
			svr.Ports = ports
		}

		if len(service.Volumes) != 0 {
			volumes := []string{}
			volumes = append(volumes, service.Volumes...)
			svr.Volumes = volumes
		}

		if service.Options != "" {
			svr.Options = service.Options
		}

		services[service.Name] = svr
	}

	job := Job{
		Services: services,
	}

	return job, nil
}
