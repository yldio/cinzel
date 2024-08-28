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

	var name1 = "staging_environment"
	var name2 = "production_environment"
	var url2 = "${{ steps.step_id.outputs.url_output }}"

	var have1 = EnvironmentConfig{
		Name: &name1,
	}
	var expect1 = name1

	var have2 = EnvironmentConfig{
		Name: &name2,
		Url:  &url2,
	}
	var expect2 = map[string]any{
		"name": "production_environment",
		"url":  "${{ steps.step_id.outputs.url_output }}",
	}

	var tests = []Test{
		{"with defined environment", &have1, expect1},
		{"without empty environment", &have2, expect2},
		{"without undefined environment", nil, nil},
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
