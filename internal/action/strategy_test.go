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

	matrix_name_1 := "os"
	matrix_value_1_1 := cty.StringVal("ubuntu-latest")
	matrix_value_1_2 := cty.StringVal("windows-latest")
	matrix_value_1 := []*cty.Value{
		&matrix_value_1_1,
		&matrix_value_1_2,
	}

	matrix_name_2 := "version"
	matrix_value_2_1 := cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(10))
	matrix_value_2_2 := cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(12))
	matrix_value_2_3 := cty.NumberVal(new(big.Float).SetPrec(512).SetInt64(14))
	matrix_value_2 := []*cty.Value{
		&matrix_value_2_1,
		&matrix_value_2_2,
		&matrix_value_2_3,
	}

	matrix_name_3 := "site"
	matrix_value_3 := cty.StringVal("production")
	matrix_name_4 := "datacenter"
	matrix_value_4 := cty.StringVal("site-a")

	var have_1 = StrategyConfig{
		Matrix: MatrixesConfig{
			{
				Name:  &matrix_name_1,
				Value: &matrix_value_1,
			},
			{
				Name:  &matrix_name_2,
				Value: &matrix_value_2,
			},
			{
				Include: []*MatrixPropConfig{
					{
						Name:  &matrix_name_3,
						Value: &matrix_value_3,
					},
					{
						Name:  &matrix_name_4,
						Value: &matrix_value_4,
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
	var expect_1 = Strategy{
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
		{"with defined strategy", &have_1, expect_1},
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
