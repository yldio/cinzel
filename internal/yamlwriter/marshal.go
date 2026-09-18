// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil, nil
		}

		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		// A cty.Value is a struct of unexported fields, so the walk below reads
		// nothing off it and writes "{}". The check used to sit on the field,
		// which caught a field typed cty.Value and missed one reached through an
		// "any" field, a map value or a slice element: the value was dropped and
		// the empty map went out in its place. Checking the value catches every
		// route in.
		if val.Type() == reflect.TypeOf(cty.Value{}) {
			return convertCty(val.Interface().(cty.Value))
		}

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

			convertedValue, err := convert(field)
			if err != nil {
				return nil, err
			}

			if yamlTag != "" && convertedValue != nil {
				result[yamlTag] = convertedValue
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

// convertCty renders a cty.Value as the plain Go value YAML wants, or nil when
// there is nothing to write.
//
// cty.Value requires go-cty-yaml for marshaling (it understands the cty type
// system); the result is then decoded into plain Go types via go-yaml so the
// rest of the pipeline handles it uniformly.
func convertCty(val cty.Value) (any, error) {
	if !val.IsKnown() || val.IsNull() {
		return nil, nil
	}

	yamlBytes, err := ctyyaml.Marshal(val)
	if err != nil {
		return nil, err
	}

	var out any

	if err := yaml.Unmarshal(yamlBytes, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func stripTag(tag string) string {
	if tag == "" {
		return tag
	}

	sub := strings.Split(tag, ",")

	return sub[0]
}
