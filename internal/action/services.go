// SPDX-License-Identifier: MIT
// Copyright (c) 2024 YLD Limited

package action

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/acto/internal/actoparser"
	"github.com/yldio/acto/internal/variables"
	"github.com/zclconf/go-cty/cty"
)

type Services map[string]*Service

type Service struct {
	Name        string       `yaml:"-"`
	Image       string       `yaml:"image,omitempty" hcl:"image"`
	Credentials *Credentials `yaml:"credentials,omitempty" hcl:"credentials"`
	Env         *Envs        `yaml:"env,omitempty" hcl:"env"`
	Ports       []*string    `yaml:"ports,omitempty" hcl:"ports"`
	Volumes     []*string    `yaml:"volumes,omitempty" hcl:"volumes"`
	Options     *string      `yaml:"options,omitempty" hcl:"options"`
}

type ServicesConfig []*ServiceConfig
type ServiceConfig struct {
	Identifier  string             `hcl:"_,label"`
	Image       hcl.Expression     `hcl:"image,attr"`
	Credentials *CredentialsConfig `hcl:"credentials,block"`
	Env         EnvsConfig         `hcl:"env,block"`
	Ports       hcl.Expression     `hcl:"ports,attr"`
	Volumes     hcl.Expression     `hcl:"volumes,attr"`
	Options     hcl.Expression     `hcl:"options,attr"`
}

func (config *ServiceConfig) unwrapOptions(acto *actoparser.Acto) (*string, error) {
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

		return config.unwrapOptions(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'options' must be a string")
	}
}

func (config *ServiceConfig) parseOptions() (*string, error) {
	acto := actoparser.NewActo(config.Options)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapOptions(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *ServiceConfig) unwrapVolumes(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []string:
		list := []*string{}

		for _, val := range resultValue {
			list = append(list, &val)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapVolumes(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'volumes' must be a list of strings")
	}
}

func (config *ServiceConfig) parseVolumes() ([]*string, error) {
	acto := actoparser.NewActo(config.Volumes)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapVolumes(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *ServiceConfig) unwrapPorts(acto *actoparser.Acto) ([]*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []string:
		list := []*string{}

		for _, val := range resultValue {
			list = append(list, &val)
		}

		return list, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapPorts(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'ports' must be a list of strings")
	}
}

func (config *ServiceConfig) parsePorts() ([]*string, error) {
	acto := actoparser.NewActo(config.Ports)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapPorts(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *ServiceConfig) parseEnvs() (*Envs, error) {
	env, err := config.Env.Parse()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (config *ServiceConfig) parseCredentials() (*Credentials, error) {
	value, err := config.Credentials.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in credentials: %w", err)
	}

	return value, nil
}

func (config *ServiceConfig) unwrapImage(acto *actoparser.Acto) (*string, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, errors.New("attribute 'image' must be a string")
	case string:
		return &resultValue, nil
	case actoparser.ActoVariableRef:
		variableValue, err := variables.Instance().GetValue(resultValue.Attr, resultValue.Index)
		if err != nil {
			return nil, err
		}

		return config.unwrapImage(actoparser.NewActoFromResult(variableValue))
	default:
		return nil, errors.New("attribute 'image' must be a string")
	}
}

func (config *ServiceConfig) parseImage() (*string, error) {
	acto := actoparser.NewActo(config.Image)

	if err := acto.Parse(); err != nil {
		return nil, err
	}

	value, err := config.unwrapImage(acto)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (config *ServiceConfig) Parse() (*Service, error) {
	if config == nil {
		return nil, nil
	}

	if config.Identifier == "" {
		return nil, errors.New("error in 'service': missing 'identifier'")
	}

	service := Service{
		Name: config.Identifier,
	}

	image, err := config.parseImage()
	if err != nil {
		return nil, fmt.Errorf("error in 'service': %w", err)
	}

	service.Image = *image

	credentials, err := config.parseCredentials()
	if err != nil {
		return nil, fmt.Errorf("error in 'service': %w", err)
	}

	if credentials != nil {
		service.Credentials = credentials
	}

	env, err := config.parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("error in 'service': %w", err)
	}

	if env != nil {
		service.Env = env
	}

	ports, err := config.parsePorts()
	if err != nil {
		return nil, fmt.Errorf("error in 'service': %w", err)
	}

	if ports != nil {
		service.Ports = ports
	}

	volumes, err := config.parseVolumes()
	if err != nil {
		return nil, fmt.Errorf("error in 'Service': %w", err)
	}

	if volumes != nil {
		service.Volumes = volumes
	}

	options, err := config.parseOptions()
	if err != nil {
		return nil, fmt.Errorf("error in 'Service': %w", err)
	}

	if options != nil {
		service.Options = options
	}

	return &service, nil
}

func (services *Services) Decode(body *hclwrite.Body, attr string) error {
	for id, service := range *services {
		if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
			body.AppendNewline()
		}

		serviceBlock := body.AppendNewBlock(attr, []string{id})
		serviceBody := serviceBlock.Body()

		if service.Image != "" {
			imageAttr, err := actoparser.GetHclTag(*service, "Image")
			if err != nil {
				return err
			}

			serviceBody.SetAttributeValue(imageAttr, cty.StringVal(service.Image))
		}

		if service.Credentials != nil {
			credentialsAttr, err := actoparser.GetHclTag(*service, "Credentials")
			if err != nil {
				return err
			}

			if err := service.Credentials.Decode(serviceBody, credentialsAttr); err != nil {
				return err
			}
		}

		if service.Env != nil {
			envAttr, err := actoparser.GetHclTag(*service, "Env")
			if err != nil {
				return err
			}

			if err := service.Env.Decode(serviceBody, envAttr); err != nil {
				return err
			}
		}

		if service.Ports != nil {
			portsAttr, err := actoparser.GetHclTag(*service, "Ports")
			if err != nil {
				return err
			}

			if len(serviceBody.Blocks()) > 0 || len(serviceBody.Attributes()) > 0 {
				serviceBody.AppendNewline()
			}

			var ports []cty.Value

			for _, port := range service.Ports {
				ports = append(ports, cty.StringVal(*port))
			}

			serviceBody.SetAttributeValue(portsAttr, cty.TupleVal(ports))
		}

		if service.Volumes != nil {
			volumesAttr, err := actoparser.GetHclTag(*service, "Volumes")
			if err != nil {
				return err
			}

			if len(serviceBody.Blocks()) > 0 || len(serviceBody.Attributes()) > 0 {
				serviceBody.AppendNewline()
			}

			var volumes []cty.Value

			for _, volume := range service.Volumes {
				volumes = append(volumes, cty.StringVal(*volume))
			}

			serviceBody.SetAttributeValue(volumesAttr, cty.TupleVal(volumes))
		}

		if service.Options != nil {
			optionsAttr, err := actoparser.GetHclTag(*service, "Options")
			if err != nil {
				return err
			}

			if len(serviceBody.Blocks()) > 0 || len(serviceBody.Attributes()) > 0 {
				serviceBody.AppendNewline()
			}

			serviceBody.SetAttributeValue(optionsAttr, cty.StringVal(*service.Options))
		}
	}
	return nil
}
