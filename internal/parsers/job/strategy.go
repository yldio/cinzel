package job

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type IncludeItem struct {
	Name  string    `hcl:"name,attr"`
	Value cty.Value `hcl:"value,attr"`
}

type MatrixProp struct {
	Name  *string        `hcl:"name,attr"`
	Value *cty.Value     `hcl:"value,attr"`
	Item  []*IncludeItem `hcl:"item,block"`
}

type Matrix struct {
	Name    string        `hcl:"name,attr"`
	Value   []cty.Value   `hcl:"value,attr"`
	Include []*MatrixProp `hcl:"include,block"`
	Exclude []*MatrixProp `hcl:"exclude,block"`
}

type Matrixes []*Matrix

type StrategyConfig struct {
	Matrix      Matrixes `hcl:"matrix,block" yaml:"matrix"`
	FailFast    *bool    `hcl:"fail_fast,attr" yaml:"fail-fast,omitempty"`
	MaxParallel *uint16  `hcl:"max_parallel,attr" yaml:"max-parallel,omitempty"`
}

func (config *Matrix) Parse(matrixes map[string]any) (map[string]any, error) {
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
					err := gocty.FromCtyValue(*include.Value, &val)
					if err != nil {
						return map[string]any{}, err
					}

					switch includeMap := matrixes["include"].(type) {
					case []map[string]any:
						matrixes["include"] = append(includeMap, map[string]any{
							*include.Name: val,
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

func (config *Matrixes) Parse() (map[string]any, error) {
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

func (config *StrategyConfig) Parse() (map[string]any, error) {
	content, err := config.Matrix.Parse()
	if err != nil {
		return map[string]any{}, err
	}
	return content, nil
}
