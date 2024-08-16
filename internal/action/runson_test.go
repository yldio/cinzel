// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestRunsOn(t *testing.T) {
	type Test struct {
		name   string
		have   *RunsOnConfig
		expect any
	}

	on_1 := cty.StringVal("ubuntu-latest")
	on_2 := cty.TupleVal([]cty.Value{cty.StringVal("self-hosted"), cty.StringVal("linux")})
	on_3 := "ubuntu-runners"

	var have_1 = RunsOnConfig{
		On: &on_1,
	}
	var expect_1 = "ubuntu-latest"

	var have_2 = RunsOnConfig{
		On: &on_2,
	}
	var expect_2 = []string{"self-hosted", "linux"}

	var have_3 = RunsOnConfig{
		OnGroup: &on_3,
	}
	var expect_3 = map[string]any{
		"group": "ubuntu-runners",
	}

	var tests = []Test{
		{"with defined a single runs-on", &have_1, expect_1},
		{"without empty a multi runs-on", &have_2, expect_2},
		{"without froup runs-on", &have_3, expect_3},
		{"without runs-on", nil, nil},
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
