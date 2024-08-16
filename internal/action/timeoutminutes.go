// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type TimeoutMinutesConfig uint16

func (config *TimeoutMinutesConfig) Parse() (*uint16, error) {
	if config == nil {
		return nil, nil
	}

	var number = uint16(*config)

	return &number, nil
}
