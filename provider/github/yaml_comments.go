// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"github.com/yldio/cinzel/internal/fsutil"
	yamlv3 "gopkg.in/yaml.v3"
)

// nodeComment holds the two comments a YAML key can carry: the run written on
// its own lines above the key, and the one sharing the key's line.
type nodeComment struct {
	head string
	line string
}

// withoutHead returns the comment with its head run dropped, for a writer
// whose caller has already emitted it.
func (c nodeComment) withoutHead() nodeComment {
	c.head = ""

	return c
}

// empty reports whether the key carried no comment at all.
func (c nodeComment) empty() bool {
	return c.head == "" && c.line == ""
}

// yamlComments mirrors the shape of one YAML mapping, holding the comments
// written on its keys and the same again for each mapping nested under it.
//
// The decoded map cannot hold a comment, and wrapping the values to make it
// able to would put the wrapper in front of every validator that reads the
// document — the mistake the parse direction already made once and had to
// unpick. A tree beside the map instead leaves the document exactly as it was.
//
// Every method is nil-safe, so a document with no comments costs its writers
// nothing: they walk a nil tree and read empty comments out of it.
type yamlComments struct {
	own      map[string]nodeComment
	children map[string]*yamlComments
	// items holds the comments of each mapping in a sequence under this key,
	// by position. Steps are the only sequence the providers write, and a
	// step has no key of its own to be found under.
	items map[string][]*yamlComments
	// head is the comment written above this mapping itself, which is how a
	// sequence item carries one: it has no key for it to sit on.
	head string
}

// at returns the comments written on key, or an empty nodeComment.
func (c *yamlComments) at(key string) nodeComment {
	if c == nil {
		return nodeComment{}
	}

	return c.own[key]
}

// child returns the comments of the mapping nested under key, which is nil
// when there is no such mapping or it holds no comments.
func (c *yamlComments) child(key string) *yamlComments {
	if c == nil {
		return nil
	}

	return c.children[key]
}

// item returns the comments of the idx-th mapping in the sequence under key,
// which is nil when there is no such item or it holds no comments.
func (c *yamlComments) item(key string, idx int) *yamlComments {
	if c == nil || idx < 0 || idx >= len(c.items[key]) {
		return nil
	}

	return c.items[key][idx]
}

// above returns the comment written above this mapping, which is only ever set
// for a sequence item.
func (c *yamlComments) above() string {
	if c == nil {
		return ""
	}

	return c.head
}

// collectComments reads the comments off a mapping node, every mapping nested
// under it, and every mapping in a sequence under it, returning nil when there
// are none to read.
func collectComments(node *yamlv3.Node) *yamlComments {
	if node == nil || node.Kind != yamlv3.MappingNode {
		return nil
	}

	out := &yamlComments{}

	for i := 0; i+1 < len(node.Content); i += 2 {
		key, value := node.Content[i], node.Content[i+1]

		// The head comment is written above the key and the inline one after
		// the value, which is where yaml.v3 records each of them.
		comment := nodeComment{
			head: fsutil.WithoutGeneratedMarker(key.HeadComment),
			line: value.LineComment,
		}

		if !comment.empty() {
			if out.own == nil {
				out.own = map[string]nodeComment{}
			}

			out.own[key.Value] = comment
		}

		if nested := collectComments(value); nested != nil {
			if out.children == nil {
				out.children = map[string]*yamlComments{}
			}

			out.children[key.Value] = nested
		}

		if seq := collectSeqComments(value); seq != nil {
			if out.items == nil {
				out.items = map[string][]*yamlComments{}
			}

			out.items[key.Value] = seq
		}
	}

	if out.own == nil && out.children == nil && out.items == nil {
		return nil
	}

	return out
}

// collectSeqComments reads the comments off each mapping in a sequence,
// returning nil when no item carries one. The slice is positional, so an item
// with no comments is a nil entry rather than a missing one.
//
// A comment written above a whole item sits on the item's mapping node, since
// there is no key above it for yaml.v3 to hang it on.
func collectSeqComments(node *yamlv3.Node) []*yamlComments {
	if node == nil || node.Kind != yamlv3.SequenceNode {
		return nil
	}

	out := make([]*yamlComments, len(node.Content))
	found := false

	for i, item := range node.Content {
		comments := collectComments(item)

		if head := fsutil.WithoutGeneratedMarker(item.HeadComment); head != "" {
			if comments == nil {
				comments = &yamlComments{}
			}

			comments.head = head
		}

		if comments != nil {
			found = true
		}

		out[i] = comments
	}

	if !found {
		return nil
	}

	return out
}
