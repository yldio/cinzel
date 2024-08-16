// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestDefaults(t *testing.T) {
	type Test struct {
		name   string
		have   *DefaultsConfig
		expect Defaults
	}

	var shell = "bash"
	var workingDirectory = "./scripts"

	var have_1 = DefaultsConfig{
		Run: &DefaultsRunConfig{
			Shell:            &shell,
			WorkingDirectory: &workingDirectory,
		},
	}
	var expect_1 = Defaults{
		Run: &Run{
			Shell:            &shell,
			WorkingDirectory: &workingDirectory,
		},
	}

	var have_2 = DefaultsConfig{
		Run: &DefaultsRunConfig{
			Shell: &shell,
		},
	}
	var expect_2 = Defaults{
		Run: &Run{
			Shell: &shell,
		},
	}

	var have_3 = DefaultsConfig{
		Run: &DefaultsRunConfig{
			WorkingDirectory: &workingDirectory,
		},
	}
	var expect_3 = Defaults{
		Run: &Run{
			WorkingDirectory: &workingDirectory,
		},
	}

	var tests = []Test{
		{"with defined defaults, run shell and working-directory", &have_1, expect_1},
		{"with defined defaults, run shell", &have_2, expect_2},
		{"with defined defaults and run working-directory", &have_3, expect_3},
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
