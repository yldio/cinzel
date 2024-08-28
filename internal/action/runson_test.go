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

	on1 := cty.StringVal("ubuntu-latest")
	on2 := cty.TupleVal([]cty.Value{cty.StringVal("self-hosted"), cty.StringVal("linux")})
	on3 := "ubuntu-runners"

	var have1 = RunsOnConfig{
		On: &on1,
	}
	var expect1 = "ubuntu-latest"

	var have2 = RunsOnConfig{
		On: &on2,
	}
	var expect2 = []string{"self-hosted", "linux"}

	var have3 = RunsOnConfig{
		OnGroup: &on3,
	}
	var expect3 = map[string]any{
		"group": "ubuntu-runners",
	}

	var tests = []Test{
		{"with defined a single runs-on", &have1, expect1},
		{"without empty a multi runs-on", &have2, expect2},
		{"without froup runs-on", &have3, expect3},
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
