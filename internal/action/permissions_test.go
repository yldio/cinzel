// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestPermissions(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *PermissionsConfig
	// 	expect Permissions
	// }

	// var action1 = Read
	// var action2 = Write
	// var action3 = None

	// var have1 = PermissionsConfig{
	// 	Actions:      &action1,
	// 	Issues:       &action2,
	// 	PullRequests: &action3,
	// }
	// var expect1 = Permissions{
	// 	Actions:      &action1,
	// 	Issues:       &action2,
	// 	PullRequests: &action3,
	// }

	// var expect2 Permissions

	// var tests = []Test{
	// 	{"with defined permissions", &have1, expect1},
	// 	{"without undefined permissions", nil, expect2},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := tt.have.Parse()
	// 		if err != nil {
	// 			t.Fatal(err.Error())
	// 		}

	// 		if !reflect.DeepEqual(got, tt.expect) {
	// 			t.Fatal(tt.name)
	// 		}
	// 	})
	// }
}
