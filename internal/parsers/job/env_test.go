// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
	"github.com/zclconf/go-cty/cty"
)

func TestJobOnlyWithEnv(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  env {
    variable {
      name = "NODE_ENV"
      value = "development"
    }

    variable {
      name = "TOKEN"
      value = "$${{ secrets.token }}"
    }
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
					Env: EnvConfig{
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
				Id: "job_1",
				Env: Env{
					"NODE_ENV": "development",
					"TOKEN":    "${{ secrets.token }}",
				},
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
  env:
    NODE_ENV: development
    TOKEN: ${{ secrets.token }}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
