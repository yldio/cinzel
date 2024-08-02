// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJobOnlyWithConcurrency(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  concurrency {
    group = "$${{ github.workflow }}-$${{ github.ref }}"
    cancel_in_progress = true
  }
}

job "job_2" {
  concurrency {
    group = "$${{ github.workflow }}-$${{ github.ref }}"
  }
}

job "job_3" {}
`

		var got_hcl HclConfig

		if err := parsers.HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		group_1 := "${{ github.workflow }}-${{ github.ref }}"
		cancelInProgress_1 := true
		group_2 := "${{ github.workflow }}-${{ github.ref }}"

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Concurrency: ConcurrencyConfig{
						Group:            &group_1,
						CancelInProgress: &cancelInProgress_1,
					},
				},
				{
					Id: "job_2",
					Concurrency: ConcurrencyConfig{
						Group: &group_2,
					},
				},
				{
					Id: "job_3",
				},
			},
		}

		if !reflect.DeepEqual(got_hcl, expected_hcl) {
			t.FailNow()
		}

		got_parsed, err := got_hcl.Parse()
		if err != nil {
			t.FailNow()
		}

		expected_parsed := Jobs{
			"job_1": Job{
				Id: "job_1",
				Concurrency: Concurrency{
					Group:            &group_1,
					CancelInProgress: &cancelInProgress_1,
				},
			},
			"job_2": Job{
				Id: "job_2",
				Concurrency: Concurrency{
					Group: &group_2,
				},
			},
			"job_3": Job{
				Id: "job_3",
			},
		}

		if !reflect.DeepEqual(got_parsed, expected_parsed) {
			t.FailNow()
		}

		got_yaml, err := parsers.Convert(got_parsed)
		if err != nil {
			t.FailNow()
		}

		expected_yaml := `job_1:
  concurrency:
    group: ${{ github.workflow }}-${{ github.ref }}
    cancel-in-progress: true
job_2:
  concurrency:
    group: ${{ github.workflow }}-${{ github.ref }}
job_3: {}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
