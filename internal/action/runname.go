// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type RunNameConfig string

func (config *RunNameConfig) Parse() (string, error) {
	if config == nil {
		return "", nil
	}

	return string(*config), nil
}
