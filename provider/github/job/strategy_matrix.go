// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package job

import (
	"errors"
	"fmt"

	"github.com/yldio/cinzel/internal/maputil"
)

// MatrixVariable represents a single named variable in a strategy matrix.
type MatrixVariable struct {
	Name  string
	Value any
}

// MatrixAxis represents a single axis (name-value pair) in a strategy matrix.
type MatrixAxis struct {
	Name  string
	Value any
}

// NormalizeStrategyMatrix flattens HCL "variable" blocks into top-level matrix keys.
func NormalizeStrategyMatrix(matrix map[string]any) (map[string]any, error) {
	raw, hasVariables := matrix["variable"]
	if !hasVariables {
		return matrix, nil
	}

	variables, err := matrixVariablesFromRaw(raw)
	if err != nil {
		return nil, err
	}

	for _, variable := range variables {
		if _, exists := matrix[variable.Name]; exists {
			return nil, fmt.Errorf("strategy.matrix contains duplicate key '%s'", variable.Name)
		}

		matrix[variable.Name] = variable.Value
	}

	delete(matrix, "variable")
	return matrix, nil
}

// AxesFromMap converts a map into a sorted slice of MatrixAxis values.
func AxesFromMap(mapping map[string]any) []MatrixAxis {
	keys := maputil.SortedKeys(mapping)
	axes := make([]MatrixAxis, 0, len(keys))
	for _, key := range keys {
		axes = append(axes, MatrixAxis{Name: key, Value: mapping[key]})
	}

	return axes
}

func matrixVariablesFromRaw(raw any) ([]MatrixVariable, error) {
	toVariables := func(entry map[string]any) ([]MatrixVariable, error) {
		name, nameOK := entry["name"].(string)
		value, valueOK := entry["value"]

		if nameOK || valueOK {
			if !nameOK || name == "" {
				return nil, errors.New("strategy.matrix.variable entries require a non-empty name")
			}

			if !valueOK {
				return nil, errors.New("strategy.matrix.variable entries require a value")
			}

			return []MatrixVariable{{Name: name, Value: value}}, nil
		}

		out := make([]MatrixVariable, 0, len(entry))
		for key, item := range entry {
			out = append(out, MatrixVariable{Name: key, Value: item})
		}

		return out, nil
	}

	switch v := raw.(type) {
	case []any:
		out := make([]MatrixVariable, 0, len(v))
		for _, item := range v {
			entry, ok := maputil.ToStringAnyMap(item)
			if !ok {
				return nil, errors.New("strategy.matrix.variable entries must be objects")
			}

			normalized, err := toVariables(entry)
			if err != nil {
				return nil, err
			}

			out = append(out, normalized...)
		}

		return out, nil
	case map[string]any:
		return toVariables(v)
	default:
		return nil, errors.New("strategy.matrix.variable must be an object or list")
	}
}
