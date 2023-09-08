package job

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/zclconf/go-cty/cty"
)

func TestParseStrategy(t *testing.T) {
	t.Run("convert from hcl: strategy", func(t *testing.T) {
		have := []byte(`job {
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
      }
    }

    fail_fast = true
    max_parallel = 3
  }
}
`,
		)

		var hclConfig struct {
			Jobs []struct {
				Strategy *StrategyConfig `hcl:"strategy,block"`
			} `hcl:"job,block"`
		}

		if err := HelperConvertHcl(have, &hclConfig); err != nil {
			t.Fail()
		}

		got := *hclConfig.Jobs[0].Strategy

		number_1 := new(big.Float).SetPrec(512).SetInt64(10)
		number_2 := new(big.Float).SetPrec(512).SetInt64(12)
		number_3 := new(big.Float).SetPrec(512).SetInt64(14)

		matrixPropName_1 := "site"
		matrixPropValue_1 := cty.StringVal("production")
		matrixPropName_2 := "datacenter"
		matrixPropValue_2 := cty.StringVal("site-a")

		items := []*IncludeItem{
			{
				Name:  "color",
				Value: cty.StringVal("pink"),
			},
			{
				Name:  "animal",
				Value: cty.StringVal("cat"),
			},
		}

		failFast := true
		maxParallel := uint16(3)

		expected := StrategyConfig{
			Matrix: Matrixes{
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
						cty.NumberVal(number_1),
						cty.NumberVal(number_2),
						cty.NumberVal(number_3),
					},
				},
				{
					Include: []*MatrixProp{
						{
							Name:  &matrixPropName_1,
							Value: &matrixPropValue_1,
						},
						{
							Name:  &matrixPropName_2,
							Value: &matrixPropValue_2,
						},
						{
							Item: items,
						},
					},
				},
			},
			FailFast:    &failFast,
			MaxParallel: &maxParallel,
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("parse from hcl: strategy", func(t *testing.T) {
		number_1 := new(big.Float).SetPrec(512).SetInt64(10)
		number_2 := new(big.Float).SetPrec(512).SetInt64(12)
		number_3 := new(big.Float).SetPrec(512).SetInt64(14)

		matrixPropName_1 := "site"
		matrixPropValue_1 := cty.StringVal("production")
		matrixPropName_2 := "datacenter"
		matrixPropValue_2 := cty.StringVal("site-a")
		matrixPropValue_3 := new(big.Float).SetPrec(512).SetInt64(3)

		items := []*IncludeItem{
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
				Value: cty.NumberVal(matrixPropValue_3),
			},
			{
				Name:  "safe",
				Value: cty.BoolVal(true),
			},
		}

		failFast := true
		maxParallel := uint16(3)

		have := StrategyConfig{
			Matrix: Matrixes{
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
						cty.NumberVal(number_1),
						cty.NumberVal(number_2),
						cty.NumberVal(number_3),
					},
				},
				{
					Include: []*MatrixProp{
						{
							Name:  &matrixPropName_1,
							Value: &matrixPropValue_1,
						},
						{
							Name:  &matrixPropName_2,
							Value: &matrixPropValue_2,
						},
						{
							Item: items,
						},
					},
				},
			},
			FailFast:    &failFast,
			MaxParallel: &maxParallel,
		}

		got, err := have.Parse()
		if err != nil {
			t.Fail()
		}

		expected := map[string]any{
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
		}

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})

	t.Run("convert to yaml: timeout-minutes", func(t *testing.T) {
		have := TestingStrategy{
			Strategy: map[string]any{
				"matrix": map[string]any{
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
				"fail-safe":    true,
				"max-parallel": 3,
			},
		}

		got, err := Convert(have)
		if err != nil {
			t.Fail()
		}

		expected := []byte(`strategy:
  fail-safe: true
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
  max-parallel: 3
`,
		)

		if !reflect.DeepEqual(got, expected) {
			t.Fail()
		}
	})
}
