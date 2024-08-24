// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestSecrets(t *testing.T) {
	type TestSecrets struct {
		name   string
		have   SecretsConfig
		expect Secrets
	}

	type TestSecretsInherit struct {
		name   string
		have   SecretsInheritConfig
		expect SecretsInherit
	}

	var secret1 = SecretConfig{
		Name:  "name1",
		Value: "val1",
	}

	var secret2 = SecretConfig{
		Name:  "name2",
		Value: "true",
	}

	var have1 = SecretsConfig{
		&secret1,
	}

	var expect1 = Secrets{
		"name1": "val1",
	}

	var have2 = SecretsConfig{
		&secret1,
		&secret2,
	}

	var expect2 = Secrets{
		"name1": "val1",
		"name2": "true",
	}

	var have3 = SecretsInheritConfig("inherit")

	var expect3 = SecretsInherit("inherit")

	var testSecrets = []TestSecrets{
		{"with single secrets", have1, expect1},
		{"with multiple secrets", have2, expect2},
		{"with no secrets", nil, nil},
	}

	var testSecretsInherit = []TestSecretsInherit{
		{"with a secret inherit", have3, expect3},
		{"with no secret inherit", "", ""},
	}

	for _, tt := range testSecrets {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}

	for _, tt := range testSecretsInherit {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}

	t.Run("Secrets isNill", func(t *testing.T) {
		if expect1.IsNill() == true {
			t.Fatal()
		}

		if expect2.IsNill() == true {
			t.Fatal()
		}
	})

	t.Run("SecretsInherit isNill", func(t *testing.T) {
		if expect3.IsNill() == true {
			t.Fatal()
		}
	})
}
