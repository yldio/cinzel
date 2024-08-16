// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestPermissions(t *testing.T) {
	type Test struct {
		name   string
		have   *PermissionsConfig
		expect Permissions
	}

	var action_1 = Read
	var action_2 = Write
	var action_3 = None

	var have_1 = PermissionsConfig{
		Actions:      &action_1,
		Issues:       &action_2,
		PullRequests: &action_3,
	}
	var expect_1 = Permissions{
		Actions:      &action_1,
		Issues:       &action_2,
		PullRequests: &action_3,
	}

	var expect_2 Permissions

	var tests = []Test{
		{"with defined permissions", &have_1, expect_1},
		{"without undefined permissions", nil, expect_2},
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
