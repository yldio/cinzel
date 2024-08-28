// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestEnv(t *testing.T) {
	type Test struct {
		name   string
		have   *EnvConfig
		expect Env
	}

	var have1 = EnvConfig{
		Variable: []VariableConfig{
			{
				Name:  "NODE_ENV",
				Value: cty.StringVal("development"),
			},
			{
				Name:  "TOKEN",
				Value: cty.StringVal("${{ secrets.token }}"),
			},
		},
	}
	var expect1 = Env{
		"NODE_ENV": "development",
		"TOKEN":    "${{ secrets.token }}",
	}

	var tests = []Test{
		{"with defined env", &have1, expect1},
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
