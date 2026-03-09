// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package yamlwriter

import (
	"errors"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
)

// Marshal converts input to an intermediate representation and then marshals it to YAML bytes.
func Marshal[T any](input T) ([]byte, error) {
	converted, err := Convert(input)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(converted)
}

// Convert transforms a typed value into a plain any suitable for YAML marshaling.
func Convert[T any](input T) (any, error) {
	return convert(reflect.ValueOf(input))
}

func convert(val reflect.Value) (any, error) {
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return nil, nil
		}

		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		result := make(map[string]any)
		typ := val.Type()

		for i := range val.NumField() {
			field := val.Field(i)
			fieldType := typ.Field(i)

			yamlTag := fieldType.Tag.Get("yaml")
			yamlTag = stripTag(yamlTag)

			if yamlTag == "" {
				yamlTag = fieldType.Name
			}

			if yamlTag == "-" {
				continue
			}

			if !field.CanInterface() {
				continue
			}

			if field.Type() == reflect.TypeOf(cty.Value{}) {
				ctyVal := field.Interface().(cty.Value)

				if !ctyVal.IsKnown() || ctyVal.IsNull() {
					continue
				}

				// cty.Value requires go-cty-yaml for marshaling (it understands
				// the cty type system); the result is then decoded into plain Go
				// types via go-yaml so the rest of the pipeline handles it uniformly.
				yamlBytes, err := ctyyaml.Marshal(ctyVal)
				if err != nil {
					return nil, err
				}

				var value any

				if err := yaml.Unmarshal(yamlBytes, &value); err != nil {
					return nil, err
				}

				result[yamlTag] = value
			} else {
				convertedValue, err := convert(field)
				if err != nil {
					return nil, err
				}

				if yamlTag != "" && convertedValue != nil {
					result[yamlTag] = convertedValue
				}
			}
		}

		return result, nil

	case reflect.Slice, reflect.Array:
		var list []any

		for i := range val.Len() {
			elem, err := convert(val.Index(i))
			if err != nil {
				return nil, err
			}
			list = append(list, elem)
		}

		return list, nil

	case reflect.Map:
		result := make(map[any]any)

		for _, key := range val.MapKeys() {
			value, err := convert(val.MapIndex(key))
			if err != nil {
				return nil, err
			}
			result[key.Interface()] = value
		}

		return result, nil

	default:
		if val.CanInterface() {
			return val.Interface(), nil
		}

		return nil, errors.New("unknown error in convert")
	}
}

func stripTag(tag string) string {
	if tag == "" {
		return tag
	}

	sub := strings.Split(tag, ",")

	return sub[0]
}
