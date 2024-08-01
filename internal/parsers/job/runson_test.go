// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
	"github.com/zclconf/go-cty/cty"
)

func TestJobOnlyWithRunsOn(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  runs {
    on = "ubuntu-latest"
  }
}

job "job_2" {
  runs {
    on = ["self-hosted", "linux"]
  }
}

job "job_3" {
  runs {
    on_group = "ubuntu-runners"
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
					Runs: RunsConfig{
						On: cty.StringVal("ubuntu-latest"),
					},
				},
				{
					Id: "job_2",
					Runs: RunsConfig{
						On: cty.TupleVal([]cty.Value{cty.StringVal("self-hosted"), cty.StringVal("linux")}),
					},
				},
				{
					Id: "job_3",
					Runs: RunsConfig{
						OnGroup: "ubuntu-runners",
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
				Id:     "job_1",
				RunsOn: "ubuntu-latest",
			},
			"job_2": Job{
				Id:     "job_2",
				RunsOn: []string{"self-hosted", "linux"},
			},
			"job_3": Job{
				Id: "job_3",
				RunsOn: map[string]any{
					"group": "ubuntu-runners",
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
  runs-on: ubuntu-latest
job_2:
  runs-on:
  - self-hosted
  - linux
job_3:
  runs-on:
    group: ubuntu-runners
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
