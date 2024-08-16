// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type WithConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

type WithsConfig []*WithConfig

func (config *WithsConfig) Parse() (map[string]any, error) {
	if config == nil {
		return nil, nil
	}

	withs := make(map[string]any)

	for _, with := range *config {
		content, err := with.Parse()
		if err != nil {
			return nil, err
		}

		withs[content.Name] = content.Value
	}

	if len(withs) == 0 {
		return nil, nil
	}

	return withs, nil
}

func (config *WithConfig) Parse() (WithConfig, error) {
	if config == nil {
		return WithConfig{}, nil
	}

	return WithConfig{
		Name:  config.Name,
		Value: config.Value,
	}, nil
}
