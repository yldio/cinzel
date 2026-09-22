// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/yldio/cinzel/provider/github/action"
	"github.com/zclconf/go-cty/cty"
)

// Comment holds the comments one step attribute can carry: the run written on
// its own lines above it, the one sharing its line, and, for an entry that
// came from a block of its own, the run closing that block.
type Comment struct {
	Head string
	Line string
	// Foot is the comment closing the block an entry was written as. Only an
	// env or with entry has one: every other attribute is a single line with
	// no body to close.
	Foot string
}

// Empty reports whether the attribute carried no comment at all.
func (c Comment) Empty() bool {
	return c.Head == "" && c.Line == "" && c.Foot == ""
}

// Comments holds the comments written on a step, keyed by the YAML key each
// attribute becomes, plus the one written above the step block itself.
//
// Keyed by YAML key rather than HCL name because both directions already speak
// in YAML keys at the point they read this: parse is about to build the YAML
// mapping, and unparse has just decoded one.
type Comments struct {
	Head string
	// Foot is the comment written at the end of the step block with nothing
	// after it. It belongs to the block rather than to any one attribute.
	Foot  string
	Attrs map[string]Comment
	// Nested holds the comments on the entries of a step's "env" and "with"
	// maps, keyed by the map's own key and then by the entry's.
	Nested map[string]map[string]Comment
}

// At returns the comments written on the attribute emitted as key.
func (c Comments) At(key string) Comment {
	return c.Attrs[key]
}

// Nest returns the comments on the entries of the map emitted as key.
func (c Comments) Nest(key string) map[string]Comment {
	return c.Nested[key]
}

// setNested records the comments found on each env or with entry under key,
// keyed by the entry name they were found for.
//
// An entry is written as a block, so it has two places a run of comments can
// sit above it: above the block, and above the value inside it. Only the
// second used to be read, so a note written above an "env {" was dropped while
// the same note on a job's env block was kept. Both are the author's, and both
// belong above the one key the block becomes, so they are joined in the order
// they were written.
func (c *Comments) setNested(hv *hclparser.HCLVars, key string, ranges map[string]hcl.Range, blocks map[string]action.BlockComment) {
	entries := map[string]Comment{}

	for name, r := range ranges {
		block := blocks[name]
		comment := Comment{
			Head: joinComments(block.Head, hv.HeadComment(r)),
			Line: hv.TrailingComment(r),
			Foot: block.Foot,
		}

		if comment.Empty() {
			continue
		}

		entries[name] = comment
	}

	if len(entries) == 0 {
		return
	}

	if c.Nested == nil {
		c.Nested = map[string]map[string]Comment{}
	}

	c.Nested[key] = entries
}

// joinComments runs two comment runs together, dropping an empty one. Each is
// already a run of whole lines, so one newline between them reads as one run.
func joinComments(above, below string) string {
	switch {
	case above == "":
		return below
	case below == "":
		return above
	default:
		return above + "\n" + below
	}
}

// set records the comments found at r under the YAML key, keeping nothing when
// there were none.
func (c *Comments) set(hv *hclparser.HCLVars, key string, r hcl.Range) {
	comment := Comment{Head: hv.HeadComment(r), Line: hv.TrailingComment(r)}

	if comment.Empty() {
		return
	}

	if c.Attrs == nil {
		c.Attrs = map[string]Comment{}
	}

	c.Attrs[key] = comment
}

// setExpr records the comments written on an optional attribute.
//
// An attribute the author left out still decodes to a non-nil expression whose
// range is the block header, so reading a comment off one takes the comment
// written above the block and files it under an attribute that is not there.
// The parsed value is what says the attribute was written at all.
func (c *Comments) setExpr(hv *hclparser.HCLVars, key string, expr hcl.Expression, value cty.Value) {
	if expr == nil || value == cty.NilVal {
		return
	}

	c.set(hv, key, expr.Range())
}

// blockComments returns the comments written above the step block whose body
// this is and at the end of it.
//
// The decode hands back a body and not the block header, so the line to look
// above is the body's own start: that is the open brace, which shares a line
// with the block type. The foot is the run above the closing brace, which is
// where the body ends.
func blockComments(body hcl.Body, hv *hclparser.HCLVars) (head, foot string) {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {
		return "", ""
	}

	return hv.HeadComment(sb.SrcRange), hv.HeadComment(sb.EndRange)
}
