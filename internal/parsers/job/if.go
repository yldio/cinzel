// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type IfConfig string

func (config *IfConfig) Parse() (string, error) {
	return string(*config), nil
}
