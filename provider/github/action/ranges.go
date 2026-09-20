// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

// ValueRanges returns the source range of each env block's value, keyed by the
// name the block resolves to.
//
// Parse returns a cty object, which cannot carry where any of it was written.
// The ranges are what a caller needs to find the comments on those entries, and
// reading them separately keeps Parse returning the value alone.
func (config *EnvListConfig) ValueRanges(hv *hclparser.HCLVars) map[string]hcl.Range {
	if config == nil {
		return nil
	}

	out := map[string]hcl.Range{}

	for _, e := range *config {
		addValueRange(out, hv, e.Name, e.Value)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// ValueRanges is EnvListConfig.ValueRanges for with blocks, which have the same
// name/value shape.
func (config *WithListConfig) ValueRanges(hv *hclparser.HCLVars) map[string]hcl.Range {
	if config == nil {
		return nil
	}

	out := map[string]hcl.Range{}

	for _, w := range *config {
		addValueRange(out, hv, w.Name, w.Value)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// addValueRange records value's range under the name expression's resolved
// string, skipping a block whose name does not resolve to one. Such a block
// fails the surrounding Parse, which reports it with the detail this pass does
// not collect.
func addValueRange(out map[string]hcl.Range, hv *hclparser.HCLVars, name, value hcl.Expression) {
	if name == nil || value == nil {
		return
	}

	hp := hclparser.New(name, hv)

	if err := hp.Parse(); err != nil {
		return
	}

	resolved := hp.Result()

	if resolved == cty.NilVal || resolved.IsNull() || !resolved.IsKnown() || resolved.Type() != cty.String {
		return
	}

	out[resolved.AsString()] = value.Range()
}

// ValueWritten reports whether the value attribute was written at all.
//
// gohcl fills an absent hcl.Expression field with a synthetic null whose range
// is empty, so "value = null" and no value at all both resolve to cty.NilVal
// and only the range tells them apart. A YAML "V:" under env or with is a name
// GitHub defines as empty rather than one it leaves out, so the null it
// becomes has to survive the trip back.
func ValueWritten(expr hcl.Expression) bool {
	return expr != nil && !expr.Range().Empty()
}
