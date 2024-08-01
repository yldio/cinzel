// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
)

func TestJobOnlyWithUses(t *testing.T) {

	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  uses = "./.github/workflows/workflow-2.yml"
}
`

		var got_hcl HclConfig

		if err := parsers.HelperConvertHcl([]byte(have_hcl), &got_hcl); err != nil {
			t.FailNow()
		}

		expected_hcl := HclConfig{
			Jobs: JobsConfig{
				{
					Id:   "job_1",
					Uses: UsesConfig("./.github/workflows/workflow-2.yml"),
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
				Id:   "job_1",
				Uses: Uses("./.github/workflows/workflow-2.yml"),
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
  uses: ./.github/workflows/workflow-2.yml
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
