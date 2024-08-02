// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type Run struct {
	Shell            *string `yaml:"shell,omitempty"`
	WorkingDirectory *string `yaml:"working-directory,omitempty"`
}

type Defaults struct {
	Run *Run `yaml:"run,omitempty"`
}

type RunConfig struct {
	Shell            *string `hcl:"shell,attr"`
	WorkingDirectory *string `hcl:"working_directory,attr"`
}

type DefaultsConfig struct {
	Run *RunConfig `hcl:"run,block"`
}

func (config *DefaultsConfig) Parse() (Defaults, error) {
	if config == nil {
		return Defaults{}, nil
	}

	defaults := Defaults{}

	if config.Run == nil {
		return Defaults{}, nil
	}

	defaults.Run = &Run{}

	if config.Run.Shell != nil {
		defaults.Run.Shell = config.Run.Shell
	}

	if config.Run.WorkingDirectory != nil {
		defaults.Run.WorkingDirectory = config.Run.WorkingDirectory
	}

	return defaults, nil
}
