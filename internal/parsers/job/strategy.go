// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IncludeItemConfig struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type MatrixPropConfig struct {
	Name  string              `hcl:"name,attr"`
	Value cty.Value           `hcl:"value,attr"`
	Item  []IncludeItemConfig `hcl:"item,block"`
}

type MatrixConfig struct {
	Name    string             `hcl:"name,attr"`
	Value   []cty.Value        `hcl:"value,attr"`
	Include []MatrixPropConfig `hcl:"include,block"`
	Exclude []MatrixPropConfig `hcl:"exclude,block"`
}

type MatrixesConfig []MatrixConfig

type StrategyConfig struct {
	Matrix      MatrixesConfig `hcl:"matrix,block"`
	FailFast    bool           `hcl:"fail_fast,attr"`
	MaxParallel uint16         `hcl:"max_parallel,attr"`
}

type Matrixes map[string]any

type Strategy struct {
	Matrix      Matrixes `yaml:"matrix,omitempty"`
	FailFast    bool     `yaml:"fail-fast,omitempty"`
	MaxParallel uint16   `yaml:"max-parallel,omitempty"`
}

func (config *MatrixConfig) Parse(matrixes map[string]any) (map[string]any, error) {
	if config.Include != nil {
		if matrixes["include"] == nil {
			matrixes["include"] = []map[string]any{}
		}

		for _, include := range config.Include {
			if include.Item != nil {
				items := map[string]any{}
				for _, item := range include.Item {
					switch item.Value.Type().FriendlyName() {
					case "string":
						var val string
						err := gocty.FromCtyValue(item.Value, &val)
						if err != nil {
							return map[string]any{}, err
						}
						items[item.Name] = val
					case "number":
						var val int32
						err := gocty.FromCtyValue(item.Value, &val)
						if err != nil {
							var val float32
							err := gocty.FromCtyValue(item.Value, &val)
							if err != nil {
								return map[string]any{}, err
							}
						}
						items[item.Name] = val
					case "bool":
						var val bool
						err := gocty.FromCtyValue(item.Value, &val)
						if err != nil {
							return map[string]any{}, err
						}
						items[item.Name] = val
					}
				}

				switch list := matrixes["include"].(type) {
				case []map[string]any:
					matrixes["include"] = append(list, items)
				}
			} else {
				switch include.Value.Type().FriendlyName() {
				case "string":
					var val string
					err := gocty.FromCtyValue(include.Value, &val)
					if err != nil {
						return map[string]any{}, err
					}

					switch includeMap := matrixes["include"].(type) {
					case []map[string]any:
						matrixes["include"] = append(includeMap, map[string]any{
							include.Name: val,
						})
					default:
						fmt.Println("something wrong")
					}

				default:
					fmt.Println("something wrong")

				}
			}
		}
	}

	if config.Exclude != nil {
		for _, exclude := range config.Exclude {
			fmt.Println(exclude.Name, exclude.Value)
		}
	}

	for _, value := range config.Value {
		ctyValue := cty.Value(value)

		switch ctyValue.Type().FriendlyName() {
		case "string":
			var val string

			err := gocty.FromCtyValue(ctyValue, &val)
			if err != nil {
				return map[string]any{}, err
			}

			if matrixes[config.Name] == nil {
				matrixes[config.Name] = []string{}
			}

			switch v := matrixes[config.Name].(type) {
			case []string:
				matrixes[config.Name] = append(v, val)
			default:
				return map[string]any{}, fmt.Errorf("type must be a string")
			}
		case "bool":
			var val bool

			err := gocty.FromCtyValue(ctyValue, &val)
			if err != nil {
				return map[string]any{}, err
			}

			if matrixes[config.Name] == nil {
				matrixes[config.Name] = []bool{}
			}

			switch v := matrixes[config.Name].(type) {
			case []bool:
				matrixes[config.Name] = append(v, val)
			default:
				return map[string]any{}, fmt.Errorf("type must be a boolean")
			}
		case "number":
			var val int32

			err := gocty.FromCtyValue(ctyValue, &val)
			if err != nil {
				var val float32

				err := gocty.FromCtyValue(ctyValue, &val)
				if err != nil {
					return map[string]any{}, err
				}

				if matrixes[config.Name] == nil {
					matrixes[config.Name] = []any{}
				}

				switch v := matrixes[config.Name].(type) {
				case []float32:
					matrixes[config.Name] = append(v, val)
				default:
					return map[string]any{}, fmt.Errorf("type must be a float")
				}
			} else {
				if matrixes[config.Name] == nil {
					matrixes[config.Name] = []int32{}
				}

				switch v := matrixes[config.Name].(type) {
				case []int32:
					matrixes[config.Name] = append(v, val)
				default:
					return map[string]any{}, fmt.Errorf("type must be an integer")
				}
			}
		default:
			fmt.Println("don't know")
		}
	}
	return matrixes, nil
}

func (config *MatrixesConfig) Parse() (Matrixes, error) {
	matrixes := make(map[string]any)
	for _, matrix := range *config {
		accumulator, err := matrix.Parse(matrixes)
		if err != nil {
			return map[string]any{}, err
		}
		matrixes = accumulator
	}
	return matrixes, nil
}

func (config *StrategyConfig) Parse() (Strategy, error) {
	matrix, err := config.Matrix.Parse()
	if err != nil {
		return Strategy{}, err
	}

	strategy := Strategy{
		FailFast:    config.FailFast,
		MaxParallel: config.MaxParallel,
	}

	if len(matrix) > 0 {
		strategy.Matrix = matrix
	}

	return strategy, nil
}
