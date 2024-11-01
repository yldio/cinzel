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

type Container struct {
	Image       string       `yaml:"image,omitempty" hcl:"image"`
	Credentials *Credentials `yaml:"credentials,omitempty" hcl:"credentials"`
	Env         *Envs        `yaml:"env,omitempty" hcl:"env"`
	Ports       []*uint64    `yaml:"ports,omitempty" hcl:"ports"`
	Volumes     []*string    `yaml:"volumes,omitempty" hcl:"volumes"`
	Options     *string      `yaml:"options,omitempty" hcl:"options"`
}

type ContainerConfig struct {
	Image       hcl.Expression     `hcl:"image,attr"`
	Credentials *CredentialsConfig `hcl:"credentials,block"`
	Env         EnvsConfig         `hcl:"env,block"`
	Ports       hcl.Expression     `hcl:"ports,attr"`
	Volumes     hcl.Expression     `hcl:"volumes,attr"`
	Options     hcl.Expression     `hcl:"options,attr"`
}

func (config *ContainerConfig) unwrapOptions(acto *actoparser.Acto) (*string, error) {
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

func (config *ContainerConfig) parseOptions() (*string, error) {
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

func (config *ContainerConfig) unwrapVolumes(acto *actoparser.Acto) ([]*string, error) {
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

func (config *ContainerConfig) parseVolumes() ([]*string, error) {
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

func (config *ContainerConfig) unwrapPorts(acto *actoparser.Acto) ([]*uint64, error) {
	switch resultValue := acto.Result.(type) {
	case nil:
		return nil, nil
	case []uint64:
		list := []*uint64{}

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
		return nil, errors.New("attribute 'ports' must be a list of numbers")
	}
}

func (config *ContainerConfig) parsePorts() ([]*uint64, error) {
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

func (config *ContainerConfig) parseEnvs() (*Envs, error) {
	env, err := config.Env.Parse()
	if err != nil {
		return nil, err
	}

	return env, nil
}

func (config *ContainerConfig) parseCredentials() (*Credentials, error) {
	value, err := config.Credentials.Parse()
	if err != nil {
		return nil, fmt.Errorf("error in credentials: %w", err)
	}

	return value, nil
}

func (config *ContainerConfig) unwrapImage(acto *actoparser.Acto) (*string, error) {
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

func (config *ContainerConfig) parseImage() (*string, error) {
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

func (config *ContainerConfig) Parse() (*Container, error) {
	if config == nil {
		return nil, nil
	}

	container := Container{}

	image, err := config.parseImage()
	if err != nil {
		return nil, fmt.Errorf("error in 'container': %w", err)
	}

	container.Image = *image

	credentials, err := config.parseCredentials()
	if err != nil {
		return nil, fmt.Errorf("error in 'container': %w", err)
	}

	if credentials != nil {
		container.Credentials = credentials
	}

	env, err := config.parseEnvs()
	if err != nil {
		return nil, fmt.Errorf("error in 'container': %w", err)
	}

	if env != nil {
		container.Env = env
	}

	ports, err := config.parsePorts()
	if err != nil {
		return nil, fmt.Errorf("error in 'container': %w", err)
	}

	if ports != nil {
		container.Ports = ports
	}

	volumes, err := config.parseVolumes()
	if err != nil {
		return nil, fmt.Errorf("error in 'container': %w", err)
	}

	if volumes != nil {
		container.Volumes = volumes
	}

	options, err := config.parseOptions()
	if err != nil {
		return nil, fmt.Errorf("error in 'container': %w", err)
	}

	if options != nil {
		container.Options = options
	}

	return &container, nil
}

func (container *Container) Decode(body *hclwrite.Body, attr string) error {
	if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
		body.AppendNewline()
	}

	containerBlock := body.AppendNewBlock(attr, nil)
	containerBody := containerBlock.Body()

	if container.Image != "" {
		imageAttr, err := actoparser.GetHclTag(*container, "Image")
		if err != nil {
			return err
		}

		containerBody.SetAttributeValue(imageAttr, cty.StringVal(container.Image))
	}

	if container.Credentials != nil {
		credentialsAttr, err := actoparser.GetHclTag(*container, "Credentials")
		if err != nil {
			return err
		}

		if err := container.Credentials.Decode(containerBody, credentialsAttr); err != nil {
			return err
		}
	}

	if container.Env != nil {
		envAttr, err := actoparser.GetHclTag(*container, "Env")
		if err != nil {
			return err
		}

		if err := container.Env.Decode(containerBody, envAttr); err != nil {
			return err
		}
	}

	if container.Ports != nil {
		portsAttr, err := actoparser.GetHclTag(*container, "Ports")
		if err != nil {
			return err
		}

		if len(containerBody.Blocks()) > 0 || len(containerBody.Attributes()) > 0 {
			containerBody.AppendNewline()
		}

		var ports []cty.Value

		for _, port := range container.Ports {
			ports = append(ports, cty.NumberUIntVal(*port))
		}

		containerBody.SetAttributeValue(portsAttr, cty.TupleVal(ports))
	}

	if container.Volumes != nil {
		volumesAttr, err := actoparser.GetHclTag(*container, "Volumes")
		if err != nil {
			return err
		}

		if len(containerBody.Blocks()) > 0 || len(containerBody.Attributes()) > 0 {
			containerBody.AppendNewline()
		}

		var volumes []cty.Value

		for _, volume := range container.Volumes {
			volumes = append(volumes, cty.StringVal(*volume))
		}

		containerBody.SetAttributeValue(volumesAttr, cty.TupleVal(volumes))
	}

	if container.Options != nil {
		optionsAttr, err := actoparser.GetHclTag(*container, "Options")
		if err != nil {
			return err
		}

		if len(containerBody.Blocks()) > 0 || len(containerBody.Attributes()) > 0 {
			containerBody.AppendNewline()
		}

		containerBody.SetAttributeValue(optionsAttr, cty.StringVal(*container.Options))
	}

	return nil
}
