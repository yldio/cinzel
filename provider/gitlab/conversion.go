// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"math"
	"math/big"
	"reflect"

	"github.com/goccy/go-yaml"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
)

func ctyToAny(val cty.Value) (any, error) {
	if !val.IsKnown() || val.IsNull() {
		return nil, nil
	}

	t := val.Type()

	switch {
	case t == cty.String:
		return val.AsString(), nil
	case t == cty.Bool:
		return val.True(), nil
	case t == cty.Number:
		return ctyNumberToAny(val), nil
	case t.IsListType() || t.IsTupleType() || t.IsSetType():
		out := make([]any, 0, val.LengthInt())
		it := val.ElementIterator()

		for it.Next() {
			_, child := it.Element()
			childAny, err := ctyToAny(child)
			if err != nil {
				return nil, err
			}
			out = append(out, childAny)
		}

		return out, nil
	case t.IsMapType() || t.IsObjectType():
		out := make(map[string]any, len(val.AsValueMap()))

		for key, child := range val.AsValueMap() {
			childAny, err := ctyToAny(child)
			if err != nil {
				return nil, err
			}
			out[key] = childAny
		}

		return out, nil
	default:
		return ctyToAnyViaYAML(val)
	}
}

func anyToCty(value any) (cty.Value, error) {
	if value == nil {
		return cty.NullVal(cty.DynamicPseudoType), nil
	}

	if err := rejectInfinity(value); err != nil {
		return cty.NilVal, err
	}

	if v, ok := value.(cty.Value); ok {
		if !v.IsKnown() {
			return cty.NullVal(cty.DynamicPseudoType), nil
		}

		return v, nil
	}

	if v, ok := anyToCtyDirect(value); ok {
		return v, nil
	}

	return anyToCtyViaYAML(value)
}

func anyToCtyDirect(value any) (cty.Value, bool) {
	switch v := value.(type) {
	case string:
		return cty.StringVal(v), true
	case bool:
		return cty.BoolVal(v), true
	case int:
		return cty.NumberIntVal(int64(v)), true
	case int8:
		return cty.NumberIntVal(int64(v)), true
	case int16:
		return cty.NumberIntVal(int64(v)), true
	case int32:
		return cty.NumberIntVal(int64(v)), true
	case int64:
		return cty.NumberIntVal(v), true
	case uint:
		return cty.NumberUIntVal(uint64(v)), true
	case uint8:
		return cty.NumberUIntVal(uint64(v)), true
	case uint16:
		return cty.NumberUIntVal(uint64(v)), true
	case uint32:
		return cty.NumberUIntVal(uint64(v)), true
	case uint64:
		return cty.NumberUIntVal(v), true
	case float32:
		return ctyFloat(float64(v))
	case float64:
		return ctyFloat(v)
	case []any:
		vals := make([]cty.Value, 0, len(v))

		for _, item := range v {
			child, ok := anyToCtyDirect(item)

			if !ok {
				return cty.NilVal, false
			}
			vals = append(vals, child)
		}

		return cty.TupleVal(vals), true
	case map[string]any:
		vals := make(map[string]cty.Value, len(v))

		for key, item := range v {
			child, ok := anyToCtyDirect(item)

			if !ok {
				return cty.NilVal, false
			}
			vals[key] = child
		}

		return cty.ObjectVal(vals), true
	}

	rv := reflect.ValueOf(value)

	if !rv.IsValid() {
		return cty.NullVal(cty.DynamicPseudoType), true
	}

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return cty.NullVal(cty.DynamicPseudoType), true
		}

		return anyToCtyDirect(rv.Elem().Interface())
	}

	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		vals := make([]cty.Value, 0, rv.Len())

		for i := 0; i < rv.Len(); i++ {
			child, ok := anyToCtyDirect(rv.Index(i).Interface())

			if !ok {
				return cty.NilVal, false
			}
			vals = append(vals, child)
		}

		return cty.TupleVal(vals), true
	}

	if rv.Kind() == reflect.Map && rv.Type().Key().Kind() == reflect.String {
		vals := make(map[string]cty.Value, rv.Len())
		iter := rv.MapRange()

		for iter.Next() {
			child, ok := anyToCtyDirect(iter.Value().Interface())

			if !ok {
				return cty.NilVal, false
			}
			vals[iter.Key().String()] = child
		}

		return cty.ObjectVal(vals), true
	}

	return cty.NilVal, false
}

func ctyNumberToAny(val cty.Value) any {
	bf := val.AsBigFloat()

	if i, acc := bf.Int64(); acc == big.Exact {
		return i
	}

	f, _ := bf.Float64()

	return f
}

func ctyToAnyViaYAML(val cty.Value) (any, error) {
	bytes, err := ctyyaml.Marshal(val)
	if err != nil {
		return nil, err
	}

	var out any

	if err := yaml.Unmarshal(bytes, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// rejectInfinity refuses ".inf", at the top level or anywhere inside a list or
// a mapping.
//
// hclwrite has no token for an infinity and writes one as "+ Inf", which is not
// an HCL expression at all: the file went out with exit 0 and cinzel's own
// parse then could not read it back. actionlint refuses ".inf" outright,
// reporting `invalid float value`. A NaN is declined further down instead,
// where cty would panic on it before anything could report it.
func rejectInfinity(value any) error {
	switch v := value.(type) {
	case float32:
		return infinityError(float64(v))
	case float64:
		return infinityError(v)
	case []any:
		for _, item := range v {
			if err := rejectInfinity(item); err != nil {
				return err
			}
		}
	case map[string]any:
		for _, item := range v {
			if err := rejectInfinity(item); err != nil {
				return err
			}
		}
	}

	return nil
}

func infinityError(v float64) error {
	if math.IsInf(v, 0) {
		return errInfiniteNumber
	}

	return nil
}

// ctyFloat converts a float, declining a NaN. cty.NumberFloatVal panics on one
// rather than returning an error, and ".nan" is a float any YAML document may
// carry, so declining here sends it down the YAML path instead, which reports
// it as the unsupported value it is.
func ctyFloat(v float64) (cty.Value, bool) {
	if math.IsNaN(v) {
		return cty.NilVal, false
	}

	return cty.NumberFloatVal(v), true
}

func anyToCtyViaYAML(value any) (cty.Value, error) {
	bytes, err := yaml.Marshal(value)
	if err != nil {
		return cty.NilVal, err
	}

	val, err := ctyyaml.Unmarshal(bytes, cty.DynamicPseudoType)
	if err != nil {
		return cty.NilVal, err
	}

	if !val.IsKnown() {
		return cty.NullVal(cty.DynamicPseudoType), nil
	}

	return val, nil
}
