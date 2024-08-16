// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"reflect"
	"testing"
)

func TestWorkingDirectory(t *testing.T) {
	type Test struct {
		name   string
		have   *WorkingDirectoryConfig
		expect string
	}

	var workingDirectory = "./temp"

	var have_1 = WorkingDirectoryConfig(workingDirectory)
	var expect_1 = workingDirectory

	var have_2 = WorkingDirectoryConfig("")
	var expect_2 = ""

	var expect_3 = ""

	var tests = []Test{
		{"with defined working-directory", &have_1, expect_1},
		{"without empty working-directory", &have_2, expect_2},
		{"without undefined working-directory", nil, expect_3},
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
