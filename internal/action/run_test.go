// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestRun(t *testing.T) {
	type Test struct {
		name   string
		have   *RunConfig
		expect string
	}

	var run = `|
npm ci
npm run build
`

	var have_1 = RunConfig(run)
	var expect_1 = run

	var have_2 = RunConfig("")
	var expect_2 = ""

	var expect_3 = ""

	var tests = []Test{
		{"with defined run", &have_1, expect_1},
		{"without empty run", &have_2, expect_2},
		{"without undefined run", nil, expect_3},
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
