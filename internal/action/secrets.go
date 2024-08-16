// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

type Secrets map[string]any
type SecretsInherit string

type SecretConfig struct {
	Name  string `hcl:"name,attr"`
	Value string `hcl:"value,attr"`
}

type SecretsConfig []*SecretConfig
type SecretsInheritConfig string

func (config *SecretsConfig) Parse() (Secrets, error) {
	if config == nil {
		return nil, nil
	}
	secrets := make(Secrets)

	for _, secret := range *config {
		secrets[secret.Name] = secret.Value
	}

	if len(secrets) == 0 {
		return nil, nil
	}

	return secrets, nil
}

func (config *SecretsInheritConfig) Parse() (SecretsInherit, error) {
	if config == nil {
		return SecretsInherit(""), nil
	}

	return SecretsInherit(*config), nil
}

func (config *SecretsInherit) IsNill() bool {
	isNill := true

	if *config != "" {
		isNill = false
	}

	return isNill
}

func (config *Secrets) IsNill() bool {
	isNill := true

	if config != (&Secrets{}) {
		isNill = false
	}

	if len(*config) == 0 {
		isNill = true
	}

	return isNill
}
