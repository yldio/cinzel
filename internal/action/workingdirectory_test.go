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

	var have1 = WorkingDirectoryConfig(workingDirectory)
	var expect1 = workingDirectory

	var have2 = WorkingDirectoryConfig("")
	var expect2 = ""

	var expect3 = ""

	var tests = []Test{
		{"with defined working-directory", &have1, expect1},
		{"without empty working-directory", &have2, expect2},
		{"without undefined working-directory", nil, expect3},
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
