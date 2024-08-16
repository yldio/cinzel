// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"errors"

	"github.com/zclconf/go-cty/cty"
)

type IncludeItemConfig struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type MatrixPropConfig struct {
	Name  *string              `hcl:"name,attr"`
	Value *cty.Value           `hcl:"value,attr"`
	Items []*IncludeItemConfig `hcl:"item,block"`
}

type MatrixConfig struct {
	Name    *string             `hcl:"name,attr"`
	Value   *[]*cty.Value       `hcl:"value,attr"`
	Include []*MatrixPropConfig `hcl:"include,block"`
	Exclude []*MatrixPropConfig `hcl:"exclude,block"`
}

type MatrixesConfig []MatrixConfig

type StrategyConfig struct {
	Matrix      MatrixesConfig `hcl:"matrix,block"`
	FailFast    *bool          `hcl:"fail_fast,attr"`
	MaxParallel *uint16        `hcl:"max_parallel,attr"`
}

type Matrixes map[string]any

type Strategy struct {
	Matrix      Matrixes `yaml:"matrix,omitempty"`
	FailFast    bool     `yaml:"fail-fast,omitempty"`
	MaxParallel uint16   `yaml:"max-parallel,omitempty"`
}

func (config *MatrixConfig) Parse(matrixes map[string]any) (map[string]any, error) {
	// if config.Include != nil {
	// 	if matrixes["include"] == nil {
	// 		matrixes["include"] = []map[string]any{}
	// 	}

	// 	for _, include := range config.Include {
	// 		if include.Item != nil {
	// 			items := map[string]any{}
	// 			for _, item := range include.Item {
	// 				switch item.Value.Type().FriendlyName() {
	// 				case "string":
	// 					var val string
	// 					err := gocty.FromCtyValue(item.Value, &val)
	// 					if err != nil {
	// 						return map[string]any{}, err
	// 					}
	// 					items[item.Name] = val
	// 				case "number":
	// 					var val int32
	// 					err := gocty.FromCtyValue(item.Value, &val)
	// 					if err != nil {
	// 						var val float32
	// 						err := gocty.FromCtyValue(item.Value, &val)
	// 						if err != nil {
	// 							return map[string]any{}, err
	// 						}
	// 					}
	// 					items[item.Name] = val
	// 				case "bool":
	// 					var val bool
	// 					err := gocty.FromCtyValue(item.Value, &val)
	// 					if err != nil {
	// 						return map[string]any{}, err
	// 					}
	// 					items[item.Name] = val
	// 				}
	// 			}

	// 			switch list := matrixes["include"].(type) {
	// 			case []map[string]any:
	// 				matrixes["include"] = append(list, items)
	// 			}
	// 		} else {
	// 			switch include.Value.Type().FriendlyName() {
	// 			case "string":
	// 				var val string
	// 				err := gocty.FromCtyValue(include.Value, &val)
	// 				if err != nil {
	// 					return map[string]any{}, err
	// 				}

	// 				switch includeMap := matrixes["include"].(type) {
	// 				case []map[string]any:
	// 					matrixes["include"] = append(includeMap, map[string]any{
	// 						include.Name: val,
	// 					})
	// 				default:
	// 					fmt.Println("something wrong")
	// 				}

	// 			default:
	// 				fmt.Println("something wrong")

	// 			}
	// 		}
	// 	}
	// }

	// if config.Exclude != nil {
	// 	for _, exclude := range config.Exclude {
	// 		fmt.Println(exclude.Name, exclude.Value)
	// 	}
	// }

	// for _, value := range config.Value {
	// 	ctyValue := cty.Value(value)

	// 	switch ctyValue.Type().FriendlyName() {
	// 	case "string":
	// 		var val string

	// 		err := gocty.FromCtyValue(ctyValue, &val)
	// 		if err != nil {
	// 			return map[string]any{}, err
	// 		}

	// 		if matrixes[config.Name] == nil {
	// 			matrixes[config.Name] = []string{}
	// 		}

	// 		switch v := matrixes[config.Name].(type) {
	// 		case []string:
	// 			matrixes[config.Name] = append(v, val)
	// 		default:
	// 			return map[string]any{}, fmt.Errorf("type must be a string")
	// 		}
	// 	case "bool":
	// 		var val bool

	// 		err := gocty.FromCtyValue(ctyValue, &val)
	// 		if err != nil {
	// 			return map[string]any{}, err
	// 		}

	// 		if matrixes[config.Name] == nil {
	// 			matrixes[config.Name] = []bool{}
	// 		}

	// 		switch v := matrixes[config.Name].(type) {
	// 		case []bool:
	// 			matrixes[config.Name] = append(v, val)
	// 		default:
	// 			return map[string]any{}, fmt.Errorf("type must be a boolean")
	// 		}
	// 	case "number":
	// 		var val int32

	// 		err := gocty.FromCtyValue(ctyValue, &val)
	// 		if err != nil {
	// 			var val float32

	// 			err := gocty.FromCtyValue(ctyValue, &val)
	// 			if err != nil {
	// 				return map[string]any{}, err
	// 			}

	// 			if matrixes[config.Name] == nil {
	// 				matrixes[config.Name] = []any{}
	// 			}

	// 			switch v := matrixes[config.Name].(type) {
	// 			case []float32:
	// 				matrixes[config.Name] = append(v, val)
	// 			default:
	// 				return map[string]any{}, fmt.Errorf("type must be a float")
	// 			}
	// 		} else {
	// 			if matrixes[config.Name] == nil {
	// 				matrixes[config.Name] = []int32{}
	// 			}

	// 			switch v := matrixes[config.Name].(type) {
	// 			case []int32:
	// 				matrixes[config.Name] = append(v, val)
	// 			default:
	// 				return map[string]any{}, fmt.Errorf("type must be an integer")
	// 			}
	// 		}
	// 	default:
	// 		fmt.Println("don't know")
	// 	}
	// }
	return matrixes, nil
}

func (config *MatrixConfig) parseInclude(matrixes *Matrixes) error {
	list := []map[string]any{}

	for _, include := range config.Include {
		if include.Items != nil {
			if len(include.Items) == 0 {
				return errors.New("include items should not be empty")
			}

			ojb := make(map[string]any)

			for _, item := range include.Items {
				val, err := ParseCtyValue(item.Value, []string{
					cty.String.FriendlyName(),
					cty.Number.FriendlyName(),
					cty.Bool.FriendlyName(),
				})
				if err != nil {
					return err
				}

				ojb[item.Name] = val
			}

			list = append(list, ojb)
		} else if include.Name != nil && include.Value != nil {
			val, err := ParseCtyValue(*include.Value, []string{
				cty.String.FriendlyName(),
				cty.Number.FriendlyName(),
				cty.Bool.FriendlyName(),
			})
			if err != nil {
				return err
			}

			ojb := make(map[string]any)

			ojb[*include.Name] = val

			list = append(list, ojb)
		} else {
			return errors.New("not a valid matrix block")
		}
	}

	(*matrixes)["include"] = list

	return nil
}

func (config *MatrixConfig) parseExclude() error {
	return nil
}

func (config *MatrixConfig) parseNameValue(matrixes *Matrixes) error {
	list := []any{}

	for _, keyVal := range *config.Value {
		val, err := ParseCtyValue(*keyVal, []string{
			cty.String.FriendlyName(),
			cty.Number.FriendlyName(),
			cty.Bool.FriendlyName(),
		})
		if err != nil {
			return err
		}

		list = append(list, val)
	}

	(*matrixes)[*config.Name] = list

	return nil
}

func (config *MatrixesConfig) Parse() (Matrixes, error) {
	if config == nil {
		return nil, nil
	}
	matrixes := make(Matrixes)

	for _, matrix := range *config {
		if matrix.Include != nil {
			if err := matrix.parseInclude(&matrixes); err != nil {
				return nil, err
			}
		} else if matrix.Exclude != nil {
			if err := matrix.parseExclude(); err != nil {
				return nil, err
			}
		} else if matrix.Name != nil && matrix.Value != nil {
			if err := matrix.parseNameValue(&matrixes); err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("invalid matrix")
		}
	}
	return matrixes, nil
}

func (config *StrategyConfig) Parse() (Strategy, error) {
	if config == nil {
		return Strategy{}, nil
	}

	matrix, err := config.Matrix.Parse()
	if err != nil {
		return Strategy{}, err
	}

	if len(matrix) == 0 {
		return Strategy{}, errors.New("strategy matrix cannot be empty")
	}

	strategy := Strategy{
		Matrix: matrix,
	}

	if config.FailFast != nil {
		strategy.FailFast = *config.FailFast
	}

	if config.MaxParallel != nil {
		strategy.MaxParallel = *config.MaxParallel
	}

	return strategy, nil
}
