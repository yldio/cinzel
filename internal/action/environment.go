// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type EnvironmentConfig struct {
	Name *string `hcl:"name,attr"`
	Url  *string `hcl:"url,attr"`
}

func (config *EnvironmentConfig) Parse() (any, error) {
	if config == nil {
		return nil, nil
	}

	if config.Name == nil && config.Url == nil {
		return nil, nil
	}

	if config.Name != nil && config.Url == nil {
		return *config.Name, nil
	}

	environment := map[string]any{
		"name": *config.Name,
		"url":  *config.Url,
	}

	return environment, nil
}
