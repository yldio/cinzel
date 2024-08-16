// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type Credentials struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type CredentialsConfig struct {
	Username string `hcl:"username,attr"`
	Password string `hcl:"password,attr"`
}
