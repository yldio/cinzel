// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJobOnlyWithDefaults(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  defaults {
    run {
      shell = "bash"
      working_directory = "./scripts"
    }
  }
}

job "job_2" {
  defaults {
    run {
      shell = "bash"
    }
  }
}

job "job_3" {}
`

		var got_hcl HclConfig

		if err := parsers.HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		shell := "bash"
		workingDirectory := "./scripts"

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Defaults: DefaultsConfig{
						Run: &RunConfig{
							Shell:            &shell,
							WorkingDirectory: &workingDirectory,
						},
					},
				},
				{
					Id: "job_2",
					Defaults: DefaultsConfig{
						Run: &RunConfig{
							Shell: &shell,
						},
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
				Defaults: Defaults{
					Run: &Run{
						Shell:            &shell,
						WorkingDirectory: &workingDirectory,
					},
				},
			},
			"job_2": Job{
				Id: "job_2",
				Defaults: Defaults{
					Run: &Run{
						Shell: &shell,
					},
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
  defaults:
    run:
      shell: bash
      working-directory: ./scripts
job_2:
  defaults:
    run:
      shell: bash
job_3: {}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
