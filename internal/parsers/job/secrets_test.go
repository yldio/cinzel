// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJobOnlyWithSecrets(t *testing.T) {

	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  secret {
    name = "access-token"
    value = "$${{ secrets.PERSONAL_ACCESS_TOKEN }}"
  }

  secret {
    name = "password"
    value = "$${{ secrets.PASSWORD }}"
  }
}

job "job_2" {
  secrets = "inherit"
}
`

		var got_hcl HclConfig

		if err := HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Secret: SecretsConfig{
						{
							Name:  "access-token",
							Value: "${{ secrets.PERSONAL_ACCESS_TOKEN }}",
						},
						{
							Name:  "password",
							Value: "${{ secrets.PASSWORD }}",
						},
					},
				},
				{
					Id:      "job_2",
					Secrets: "inherit",
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
				Secrets: Secrets{
					"access-token": "${{ secrets.PERSONAL_ACCESS_TOKEN }}",
					"password":     "${{ secrets.PASSWORD }}",
				},
			},
			"job_2": Job{
				Id:      "job_2",
				Secrets: SecretsInherit("inherit"),
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
  secrets:
    access-token: ${{ secrets.PERSONAL_ACCESS_TOKEN }}
    password: ${{ secrets.PASSWORD }}
job_2:
  secrets: inherit
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
