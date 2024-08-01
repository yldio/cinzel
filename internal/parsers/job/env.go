// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import "github.com/zclconf/go-cty/cty/gocty"

type EnvConfig struct {
	Variable []VariableConfig `hcl:"variable,block"`
}

type Env map[string]any

func (config *EnvConfig) Parse() (Env, error) {
	envs := make(Env)

	for _, env := range config.Variable {
		switch env.Value.Type().FriendlyName() {
		case "string":
			var val string
			err := gocty.FromCtyValue(env.Value, &val)
			if err != nil {
				return Env{}, err
			}
			envs[env.Name] = val
		case "number":
			var val int32
			err := gocty.FromCtyValue(env.Value, &val)
			if err != nil {
				var val float32
				err := gocty.FromCtyValue(env.Value, &val)
				if err != nil {
					return Env{}, err
				}
			}
			envs[env.Name] = val
		case "bool":
			var val bool
			err := gocty.FromCtyValue(env.Value, &val)
			if err != nil {
				return Env{}, err
			}
			envs[env.Name] = val
		}
	}

	return envs, nil
}
