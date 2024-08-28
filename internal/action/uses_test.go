// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestUses(t *testing.T) {
	type Test struct {
		name   string
		have   *UsesConfig
		expect string
	}

	var have1 = UsesConfig{
		Action:  "actions/checkout",
		Version: "v4",
	}
	var expect1 = "actions/checkout@v4"

	var expect3 = ""

	var tests = []Test{
		{"with defined uses", &have1, expect1},
		{"without undefined uses", nil, expect3},
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
