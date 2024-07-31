// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type UsesConfig string

type Uses string

func (config *UsesConfig) Parse() (Uses, error) {
	return Uses(*config), nil
}
