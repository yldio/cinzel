// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type Env map[string]any

type EnvConfig struct {
	Variable []VariableConfig `hcl:"variable,block"`
}

func (config *EnvConfig) Parse() (Env, error) {
	if config == nil || config.Variable == nil {
		return nil, nil
	}

	envs := make(Env)

	for _, variable := range config.Variable {
		env, err := variable.Parse()
		if err != nil {
			return Env{}, err
		}

		envs[variable.Name] = env
	}

	return envs, nil
}
