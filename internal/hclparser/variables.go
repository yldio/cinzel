// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// HCLVars is a key-value store for HCL variable values.
type HCLVars struct {
	variables map[string]cty.Value
	sources   map[string][]byte
	lexed     map[string][]hclsyntax.Token
}

// NewHCLVars creates an empty HCLVars store.
func NewHCLVars() *HCLVars {
	return &HCLVars{
		variables: map[string]cty.Value{},
	}
}

// Add stores a variable value under the given key.
func (av *HCLVars) Add(key string, value cty.Value) {
	av.variables[key] = value
}

// GetValue retrieves a variable by key, optionally indexing into a list.
func (av *HCLVars) GetValue(attr string, idx *int64) (cty.Value, error) {
	if idx == nil {
		return av.GetValueByKey(attr)
	}

	return av.GetValueByIndex(attr, *idx)
}

// GetValueByKey returns the variable value for the given key.
func (av *HCLVars) GetValueByKey(key string) (cty.Value, error) {
	value, ok := av.variables[key]

	if ok {
		return value, nil
	}

	return cty.NilVal, fmt.Errorf("%w: `%s`", errVariableNotFound, key)
}

// GetValueByIndex returns an element from a list variable by key and index.
func (av *HCLVars) GetValueByIndex(key string, idx int64) (cty.Value, error) {
	value, err := av.GetValueByKey(key)
	if err != nil {
		return cty.NilVal, err
	}

	t := value.Type()

	// A map or an object is indexable, but by key rather than by position:
	// returning the value whole wrote the entire collection where a single
	// element was asked for. A set has no positions at all, and
	// cty.Value.Index panics on one rather than returning an error.
	if t.IsMapType() || t.IsObjectType() || t.IsSetType() {
		return cty.NilVal, fmt.Errorf("%w: %q is %s", errIndexUnsupportedType, key, t.FriendlyName())
	}

	if !t.IsListType() && !t.IsTupleType() {
		return value, nil
	}

	if !value.IsKnown() || value.IsNull() {
		return cty.NilVal, fmt.Errorf("%w: %q", errVariableNullOrUnknown, key)
	}

	length := int64(value.LengthInt())

	if idx < 0 || idx >= length {
		return cty.NilVal, fmt.Errorf("%w: %d for variable %q (length %d)", errIndexOutOfRange, idx, key, length)
	}

	return value.Index(cty.NumberIntVal(idx)), nil
}
