// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
	"github.com/zclconf/go-cty/cty"
)

func TestJobOnlyWithOutputs(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  output {
    name = "output1"
    value = "$${{ steps.step1.outputs.test }}"
  }

  output {
    name = "output2"
    value = "$${{ steps.step2.outputs.test }}"
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
					Outputs: OutputsConfig{
						{
							Name:  "output1",
							Value: cty.StringVal("${{ steps.step1.outputs.test }}"),
						},
						{
							Name:  "output2",
							Value: cty.StringVal("${{ steps.step2.outputs.test }}"),
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
				Outputs: Outputs{
					"output1": "${{ steps.step1.outputs.test }}",
					"output2": "${{ steps.step2.outputs.test }}",
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
  outputs:
    output1: ${{ steps.step1.outputs.test }}
    output2: ${{ steps.step2.outputs.test }}
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
