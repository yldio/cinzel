// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/yldio/cinzel/internal/hclparser"
)

// nodeComment holds the two comments an HCL attribute can carry: the run
// written on its own lines above it, and the one sharing its line.
type nodeComment struct {
	head string
	line string
}

// empty reports whether the attribute carried no comment at all.
func (c nodeComment) empty() bool { return c.head == "" && c.line == "" }

// comments mirrors the shape of one HCL body, holding the comments written in
// it and the same again for each body nested inside.
//
// A tree beside the pipeline map rather than a wrapper around its values: the
// map is read by validatePipeline, remapJobRefs and the emitter's own type
// switch, and a wrapper would sit in front of all three.
//
// Every method is nil-safe, so a pipeline with no comments costs its writers
// nothing: they walk a nil tree and read empty comments out of it.
type comments struct {
	own      map[string]nodeComment
	children map[string]*comments
	// items holds the comments of each body in this list, by position, for a
	// block type written more than once. Whether a block type becomes a
	// mapping or a list is the parser's call and not visible here, so a node
	// carries both and the emitter reads whichever matches the value it holds.
	items []*comments
	// head is the comment written above the block this body belongs to, which
	// is how a list item carries one: it has no key for it to sit on.
	head string
	// foot is the comment written at the end of this body with nothing after
	// it. It belongs to the body rather than to any attribute in it.
	foot string
}

// at returns the comments written on key, or an empty nodeComment.
func (c *comments) at(key string) nodeComment {
	if c == nil {
		return nodeComment{}
	}

	return c.own[key]
}

// child returns the comments of the body nested under key, nil when there is
// no such body or it holds none.
func (c *comments) child(key string) *comments {
	if c == nil {
		return nil
	}

	return c.children[key]
}

// item returns the comments of the idx-th body in this list, nil when there is
// no such item or it holds none.
func (c *comments) item(idx int) *comments {
	if c == nil || idx < 0 || idx >= len(c.items) {
		return nil
	}

	return c.items[idx]
}

// above returns the comment written above the block this body belongs to.
func (c *comments) above() string {
	if c == nil {
		return ""
	}

	return c.head
}

// below returns the comment written at the end of this body.
func (c *comments) below() string {
	if c == nil {
		return ""
	}

	return c.foot
}

// setOwn records the comments written on one key of this body, growing the map
// only when there is something to record.
func (c *comments) setOwn(key string, comment nodeComment) {
	if comment.empty() {
		return
	}

	if c.own == nil {
		c.own = map[string]nodeComment{}
	}

	c.own[key] = comment
}

// set records the comments of a child body under key, growing the tree only
// when there is something to record.
func (c *comments) set(key string, child *comments) {
	if child == nil || child.empty() {
		return
	}

	if c.children == nil {
		c.children = map[string]*comments{}
	}

	c.children[key] = child
}

// appendItem records the comments of the next body in this list. The slice is
// positional, so a body with no comments is a nil entry rather than a missing
// one.
func (c *comments) appendItem(item *comments) {
	c.items = append(c.items, item)
}

// empty reports whether nothing at all was collected.
func (c *comments) empty() bool {
	return c.own == nil && c.children == nil && c.items == nil && c.head == "" && c.foot == ""
}

// yamlKeys renames the HCL spellings that do not match the YAML key the value
// lands under. Everything absent from here is written under its own name.
//
// A "need" or "rule" block is singular because one block is one entry; the
// YAML key holding the list of them is plural.
var yamlKeys = map[string]string{
	"depends_on": "needs",
	"need":       "needs",
	"rule":       "rules",
	"service":    "services",
	"variable":   "variables",
}

// yamlKey returns the YAML key an HCL attribute or block name is written
// under.
func yamlKey(name string) string {
	if renamed, found := yamlKeys[name]; found {
		return renamed
	}

	return name
}

// topLevelSchema names everything the configuration can hold at its top level,
// which is how the comments there are reached when the input is a directory.
//
// A directory is parsed into a merged body rather than a syntax body, so the
// attributes and block headers are not there to walk; asking for them by name
// is. Each block's own body is a syntax body again, so this is needed once.
var topLevelSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{Name: "stages"}, {Name: "image"}, {Name: "before_script"},
		{Name: "after_script"}, {Name: "cache"}, {Name: "services"},
	},
	Blocks: []hcl.BlockHeaderSchema{
		{Type: "variable", LabelNames: []string{"id"}},
		{Type: "job", LabelNames: []string{"id"}},
		{Type: "template", LabelNames: []string{"id"}},
		{Type: "workflow"}, {Type: "include"}, {Type: "default"}, {Type: "spec"},
	},
}

// readTopLevelComments returns the comments written at the top level of the
// configuration, reading it through a schema so a merged body works too.
//
// There is no foot here. A block's foot is the run above its closing brace,
// and the top level of a pipeline is the file itself rather than a block, so
// a comment at the end of the file closes nothing the emitter writes. A
// directory holds several files besides, each with an end of its own.
func readTopLevelComments(body hcl.Body, hv *hclparser.HCLVars) *comments {
	content, _, diags := body.PartialContent(topLevelSchema)

	// Left to the full decode, which reports it with the detail this pass
	// deliberately does not collect.
	if diags.HasErrors() {
		return nil
	}

	out := &comments{}

	for name, attr := range content.Attributes {
		out.setOwn(yamlKey(name), nodeComment{
			head: hv.HeadComment(attr.Range),
			line: hv.TrailingComment(attr.Range),
		})
	}

	for _, block := range content.Blocks {
		// A header rather than the body's own start, because a header is what
		// survives the merge of every file in the directory with its source
		// range intact.
		addBlockComments(out, block.Type, block.Labels, block.Body, hv.HeadComment(block.DefRange), hv)
	}

	if out.empty() {
		return nil
	}

	return out
}

// readBodyComments returns the comments written in an HCL body, nil when there
// are none. Attributes and nested blocks are read the same way, because both
// become a key in the mapping the body is written out as.
func readBodyComments(body hcl.Body, hv *hclparser.HCLVars) *comments {
	sb, ok := body.(*hclsyntax.Body)

	if !ok {
		return nil
	}

	out := &comments{foot: hv.HeadComment(sb.EndRange)}

	for name, attr := range sb.Attributes {
		// "id" names the YAML key the block is written under rather than a
		// key of its own, so a comment on it belongs to the block.
		if name == "id" {
			continue
		}

		comment := nodeComment{
			head: hv.HeadComment(attr.SrcRange),
			line: hv.TrailingComment(attr.SrcRange),
		}

		out.setOwn(yamlKey(name), comment)
	}

	readNestedComments(out, sb, hv)

	if out.empty() {
		return nil
	}

	return out
}

// readNestedComments records the comments of every block nested in a body.
func readNestedComments(out *comments, body *hclsyntax.Body, hv *hclparser.HCLVars) {
	for _, block := range body.Blocks {
		addBlockComments(out, block.Type, block.Labels, block.Body, hv.HeadComment(block.TypeRange), hv)
	}
}

// addBlockComments records one block's comments under the key it is written
// out as.
//
// A labelled block becomes a mapping keyed by its label. An unlabelled one
// becomes either a mapping of its own or one entry in a list, depending on how
// many times its type was written and on what the parser does with it, so both
// readings are recorded and the emitter takes whichever fits the value.
func addBlockComments(out *comments, blockType string, labels []string, body hcl.Body, head string, hv *hclparser.HCLVars) {
	key := yamlKey(blockType)
	child := readBodyComments(body, hv)

	if head != "" {
		if child == nil {
			child = &comments{}
		}

		child.head = head
	}

	nested := out.child(key)

	if nested == nil {
		nested = &comments{}
	}

	if len(labels) == 1 {
		nested.set(labels[0], child)
		out.set(key, nested)

		return
	}

	nested.appendItem(child)

	// The first block of its type is also the whole mapping when it turns out
	// to be the only one, so its comments are merged into the node standing
	// for that key.
	if len(nested.items) == 1 && child != nil {
		nested.own = child.own
		nested.children = child.children
		nested.foot = child.foot
		nested.head = child.head
	}

	out.set(key, nested)
}
