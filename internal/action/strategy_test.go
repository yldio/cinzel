// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestStrategy(t *testing.T) {
	type Test struct {
		name   string
		have   *StrategyConfig
		expect Strategy
	}

	failFast := true
	maxParallel := uint16(3)

	matrix_name1 := "os"
	matrix_value11 := cty.StringVal("ubuntu-latest")
	matrix_value12 := cty.StringVal("windows-latest")
	matrix_value1 := []*cty.Value{
		&matrix_value11,
		&matrix_value12,
	}

	matrix_name2 := "version"
	matrix_value21 := cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(10))
	matrix_value22 := cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(12))
	matrix_value23 := cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(14))
	matrix_value2 := []*cty.Value{
		&matrix_value21,
		&matrix_value22,
		&matrix_value23,
	}

	matrix_name3 := "site"
	matrix_value3 := cty.StringVal("production")
	matrix_name4 := "datacenter"
	matrix_value4 := cty.StringVal("site-a")

	var have1 = StrategyConfig{
		Matrix: MatrixesConfig{
			{
				Name:  &matrix_name1,
				Value: &matrix_value1,
			},
			{
				Name:  &matrix_name2,
				Value: &matrix_value2,
			},
			{
				Include: []*MatrixPropConfig{
					{
						Name:  &matrix_name3,
						Value: &matrix_value3,
					},
					{
						Name:  &matrix_name4,
						Value: &matrix_value4,
					},
					{
						Items: []*IncludeItemConfig{
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
		FailFast:    &failFast,
		MaxParallel: &maxParallel,
	}
	var expect1 = Strategy{
		Matrix: Matrixes{
			"os":      []any{"ubuntu-latest", "windows-latest"},
			"version": []any{int32(10), int32(12), int32(14)},
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
	}

	var tests = []Test{
		{"with defined strategy", &have1, expect1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.have.Parse()
			if err != nil {
				t.Error(err.Error())
			}

			if !reflect.DeepEqual(got, tt.expect) {
				t.Fatalf("%s - failed", tt.name)
			}
		})
	}
}
