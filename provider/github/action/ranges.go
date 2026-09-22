// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
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
// string, skipping a block whose name does not resolve to one.
func addValueRange(out map[string]hcl.Range, hv *hclparser.HCLVars, name, value hcl.Expression) {
	if value == nil {
		return
	}

	key, ok := resolvedName(hv, name)

	if !ok {
		return
	}

	out[key] = value.Range()
}

// BlockComment holds what was written above an env or with block and at the
// end of its body. The two travel together because they are two halves of one
// block's comments.
type BlockComment struct {
	Head string
	Foot string
}

// BlockComments returns the comments written above each env block and closing
// it, keyed by the name the block resolves to.
//
// The comments on the value expression are read separately, by ValueRanges.
// Both are needed: a note above the block and a note above the value are two
// places an author writes the same thought, and reading only the second lost
// the first. The job path reads both through blockComments on hclNamedBlock.
func (config *EnvListConfig) BlockComments(hv *hclparser.HCLVars) map[string]BlockComment {
	if config == nil {
		return nil
	}

	out := map[string]BlockComment{}

	for _, e := range *config {
		addBlockComment(out, hv, e.Name, e.Body)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// BlockComments is EnvListConfig.BlockComments for with blocks, which have the
// same name/value shape.
func (config *WithListConfig) BlockComments(hv *hclparser.HCLVars) map[string]BlockComment {
	if config == nil {
		return nil
	}

	out := map[string]BlockComment{}

	for _, w := range *config {
		addBlockComment(out, hv, w.Name, w.Body)
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

// addBlockComment records body's head and foot comments under the name
// expression's resolved string, keeping nothing when there were none.
//
// The decode hands back a body and not the block header, so the line to look
// above is the body's own start: that is the open brace, which shares a line
// with the block type. The foot is the run above the closing brace, which is
// where the body ends.
func addBlockComment(out map[string]BlockComment, hv *hclparser.HCLVars, name hcl.Expression, body hcl.Body) {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {
		return
	}

	comment := BlockComment{Head: hv.HeadComment(sb.SrcRange), Foot: hv.HeadComment(sb.EndRange)}

	if comment.Head == "" && comment.Foot == "" {
		return
	}

	key, ok := resolvedName(hv, name)

	if !ok {
		return
	}

	out[key] = comment
}

// resolvedName returns the string an env or with block's name resolves to,
// reporting false for a name that does not resolve to one. Such a block fails
// the surrounding Parse, which reports it with the detail these passes do not
// collect.
func resolvedName(hv *hclparser.HCLVars, name hcl.Expression) (string, bool) {
	if name == nil {
		return "", false
	}

	hp := hclparser.New(name, hv)

	if err := hp.Parse(); err != nil {
		return "", false
	}

	resolved := hp.Result()

	if resolved == cty.NilVal || resolved.IsNull() || !resolved.IsKnown() || resolved.Type() != cty.String {
		return "", false
	}

	return resolved.AsString(), true
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
