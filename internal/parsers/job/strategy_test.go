// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/yldio/atos/internal/parsers"
	"github.com/zclconf/go-cty/cty"
)

func TestJobOnlyWithStrategy(t *testing.T) {
	t.Run("convert from hcl to yaml", func(t *testing.T) {
		have_hcl := `job "job_1" {
  strategy {
    matrix {
      name = "os"
      value = ["ubuntu-latest", "windows-latest"]
    }

    matrix {
      name = "version"
      value = [10, 12, 14]
    }

    matrix {
      include {
        name = "site"
        value = "production"
      }

      include {
        name = "datacenter"
        value = "site-a"
      }

      include {
        item {
          name = "color"
          value = "pink"
        }
        
        item {
          name = "animal"
          value = "cat"
        }

        item {
          name = "count"
          value = 3
        }
        
        item {
          name = "safe"
          value = true
        }
      }
    }

    fail_fast = true
    max_parallel = 3
  }
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
					Strategy: StrategyConfig{
						Matrix: MatrixesConfig{
							{
								Name: "os",
								Value: []cty.Value{
									cty.StringVal("ubuntu-latest"),
									cty.StringVal("windows-latest"),
								},
							},
							{
								Name: "version",
								Value: []cty.Value{
									cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(10)),
									cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(12)),
									cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(14)),
								},
							},
							{
								Include: []MatrixPropConfig{
									{
										Name:  "site",
										Value: cty.StringVal("production"),
									},
									{
										Name:  "datacenter",
										Value: cty.StringVal("site-a"),
									},
									{
										Item: []IncludeItemConfig{
											{
												Name:  "color",
												Value: cty.StringVal("pink"),
											},
											{
												Name:  "animal",
												Value: cty.StringVal("cat"),
											},
											{
												Name:  "count",
												Value: cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(3)),
											},
											{
												Name:  "safe",
												Value: cty.BoolVal(true),
											},
										},
									},
								},
							},
						},
						FailFast:    true,
						MaxParallel: uint16(3),
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
				Strategy: Strategy{
					Matrix: map[string]any{
						"os":      []string{"ubuntu-latest", "windows-latest"},
						"version": []int32{10, 12, 14},
						"include": []map[string]any{
							{
								"site": "production",
							},
							{
								"datacenter": "site-a",
							},
							{
								"color":  "pink",
								"animal": "cat",
								"count":  int32(3),
								"safe":   true,
							},
						},
					},
					FailFast:    true,
					MaxParallel: 3,
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
  strategy:
    matrix:
      include:
      - site: production
      - datacenter: site-a
      - animal: cat
        color: pink
        count: 3
        safe: true
      os:
      - ubuntu-latest
      - windows-latest
      version:
      - 10
      - 12
      - 14
    fail-fast: true
    max-parallel: 3
`

		if !reflect.DeepEqual(got_yaml, []byte(expected_yaml)) {
			t.FailNow()
		}
	})
}
