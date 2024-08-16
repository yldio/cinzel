// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestEnvironment(t *testing.T) {
	type Test struct {
		name   string
		have   *EnvironmentConfig
		expect any
	}

	var name_1 = "staging_environment"
	var name_2 = "production_environment"
	var url_2 = "${{ steps.step_id.outputs.url_output }}"

	var have_1 = EnvironmentConfig{
		Name: &name_1,
	}
	var expect_1 = name_1

	var have_2 = EnvironmentConfig{
		Name: &name_2,
		Url:  &url_2,
	}
	var expect_2 = map[string]any{
		"name": "production_environment",
		"url":  "${{ steps.step_id.outputs.url_output }}",
	}

	// var expect_3 = nil

	var tests = []Test{
		{"with defined environment", &have_1, expect_1},
		{"without empty environment", &have_2, expect_2},
		// {"without undefined environment", nil, expect_3},
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
