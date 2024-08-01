// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type ContinueOnErrorConfig bool

func (config *ContinueOnErrorConfig) Parse() (bool, error) {
	return bool(*config), nil
}
