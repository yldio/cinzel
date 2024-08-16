// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type ContinueOnErrorConfig bool

func (config *ContinueOnErrorConfig) Parse() (*bool, error) {
	if config == nil {
		return nil, nil
	}

	boolean := bool(*config)

	return &boolean, nil
}
