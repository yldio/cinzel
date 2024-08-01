// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJobOnlyWithEnvironment(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  environment {
    name = "staging_environment"
  }
}

job "job_2" {
  environment {
    name = "production_environment"
    url = "$${{ steps.step_id.outputs.url_output }}"
  }
}
`

		var got_hcl HclConfig

		if err := parsers.HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Environment: EnvironmentConfig{
						Name: "staging_environment",
					},
				},
				{
					Id: "job_2",
					Environment: EnvironmentConfig{
						Name: "production_environment",
						Url:  "${{ steps.step_id.outputs.url_output }}",
					},
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
				Id:          "job_1",
				Environment: Environment("staging_environment"),
			},
			"job_2": Job{
				Id: "job_2",
				Environment: Environment(map[string]any{
					"name": "production_environment",
					"url":  "${{ steps.step_id.outputs.url_output }}",
				}),
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
  environment: staging_environment
job_2:
  environment:
    name: production_environment
    url: ${{ steps.step_id.outputs.url_output }}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
