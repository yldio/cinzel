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

	var have_1 = RunNameConfig(runName)
	var expect_1 = runName

	var have_2 = RunNameConfig("")
	var expect_2 = ""

	var expect_3 = ""

	var tests = []Test{
		{"with defined run-name", &have_1, expect_1},
		{"without empty run-name", &have_2, expect_2},
		{"without undefined run-name", nil, expect_3},
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
