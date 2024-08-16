// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type NameConfig string

func (config *NameConfig) Parse() (string, error) {
	if config == nil {
		return "", nil
	}

	return string(*config), nil
}
