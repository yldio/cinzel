// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestRunName(t *testing.T) {
	type Test struct {
		name   string
		have   *RunNameConfig
		expect string
	}

	var runName = "Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}"

	var have1 = RunNameConfig(runName)
	var expect1 = runName

	var have2 = RunNameConfig("")
	var expect2 = ""

	var expect3 = ""

	var tests = []Test{
		{"with defined run-name", &have1, expect1},
		{"without empty run-name", &have2, expect2},
		{"without undefined run-name", nil, expect3},
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
