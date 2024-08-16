// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"
)

func TestJobs(t *testing.T) {
	type Test struct {
		name   string
		have   *JobsConfig
		expect Jobs
	}

	var have_1 = JobsConfig{
		{
			Id: "job_1",
		},
	}
	var expect_1 = Jobs{
		{
			Id: "job_1",
		},
	}

	var tests = []Test{
		{"with defined job", &have_1, expect_1},
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
