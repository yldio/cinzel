// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type Concurrency struct {
	Group            *string `yaml:"group,omitempty"`
	CancelInProgress *bool   `yaml:"cancel-in-progress,omitempty"`
}

type ConcurrencyConfig struct {
	Group            *string `hcl:"group,attr"`
	CancelInProgress *bool   `hcl:"cancel_in_progress,attr"`
}

func (config *ConcurrencyConfig) Parse() (Concurrency, error) {
	concurrency := Concurrency{}

	if config == nil {
		return concurrency, nil
	}

	if config.Group != nil {
		concurrency.Group = config.Group
	}

	if config.CancelInProgress != nil {
		concurrency.CancelInProgress = config.CancelInProgress
	}

	return concurrency, nil
}
