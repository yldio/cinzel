// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestShell(t *testing.T) {
	type TestSecrets struct {
		name   string
		have   ShellConfig
		expect string
	}

	var have1 = ShellConfig("bash")

	var expect1 = "bash"

	var tests = []TestSecrets{
		{"with shell", have1, expect1},
		{"with no shell", "", ""},
	}

	for _, tt := range tests {
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
}
