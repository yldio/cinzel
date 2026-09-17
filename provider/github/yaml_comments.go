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

// collectComments reads the comments off a mapping node and every mapping
// nested under it, returning nil when there are none to read.
//
// A comment on a sequence item is not collected. Steps are the only sequence
// the providers write and they take their own path out, so a comment picked up
// here would have nowhere to be written back to.
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
	}

	if out.own == nil && out.children == nil {
		return nil
	}

	return out
}
