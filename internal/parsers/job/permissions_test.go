// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJobOnlyWithPermissions(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  permissions {
    actions = "read"
    issues = "write"
    pull_requests = "none"
  }
}

job "job_2" {}
`

		var got_hcl HclConfig

		if err := parsers.HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		action_1 := Read
		action_2 := Write
		action_3 := None

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id: "job_1",
					Permissions: PermissionsConfig{
						Actions:      &action_1,
						Issues:       &action_2,
						PullRequests: &action_3,
					},
				},
				{
					Id: "job_2",
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
				Permissions: Permissions{
					Actions:      &action_1,
					Issues:       &action_2,
					PullRequests: &action_3,
				},
			},
			"job_2": Job{
				Id: "job_2",
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
  permissions:
    actions: read
    issues: write
    pull-requests: none
job_2: {}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
