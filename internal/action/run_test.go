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

	var have1 = RunConfig(run)
	var expect1 = run

	var have2 = RunConfig("")
	var expect2 = ""

	var expect3 = ""

	var tests = []Test{
		{"with defined run", &have1, expect1},
		{"without empty run", &have2, expect2},
		{"without undefined run", nil, expect3},
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
