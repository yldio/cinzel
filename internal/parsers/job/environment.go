// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import "errors"

type EnvironmentConfig struct {
	Name string `hcl:"name,attr"`
	Url  string `hcl:"url,attr"`
}

type Environment any

func (config *EnvironmentConfig) Parse() (Environment, error) {
	if config == nil {
		return nil, nil
	}

	if config.Name != "" && config.Url != "" {
		return map[string]any{
			"name": config.Name,
			"url":  config.Url,
		}, nil
	} else if config.Name != "" && config.Url == "" {
		return config.Name, nil
	}

	return nil, errors.New("invalid environment")
}
