// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type TimeoutMinutesConfig uint16

func (config *TimeoutMinutesConfig) Parse() (uint16, error) {
	return uint16(*config), nil
}
