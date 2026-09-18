// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/yldio/cinzel/internal/unescape"
	yamlv3 "gopkg.in/yaml.v3"
)

// Encode renders the document as YAML. Items are emitted in document order,
// inline comments are attached to their value and head comments to their key,
// a null encodes as a bare "key:" and an empty map as "key: {}". Nothing in
// the output is recovered by rewriting encoded bytes.
func Encode(d *Doc) ([]byte, error) {
	root, err := mapNode(d)
	if err != nil {
		return nil, err
	}

	setFoot(root, d.foot)

	var buf bytes.Buffer

	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(&yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{root}}); err != nil {
		return nil, err
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return unescape.Unicode(buf.Bytes()), nil
}

func mapNode(d *Doc) (*yamlv3.Node, error) {
	node := &yamlv3.Node{Kind: yamlv3.MappingNode}

	for _, it := range d.items {
		value, err := valueNode(it.value)
		if err != nil {
			return nil, err
		}

		// The head comment sits on the key, not the value: on the value it
		// is emitted inside the mapping, above its first key, rather than
		// above the key it belongs to.
		key := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: it.key, HeadComment: it.value.head}
		node.Content = append(node.Content, key, value)
	}

	return node, nil
}

// setFoot writes the comment that closes a value onto the last node inside it,
// which is where it renders at the value's own indentation and where a reader
// hands it back. On a collection node itself it renders at column 0 after the
// whole document, which is not where its author wrote it.
//
// A scalar has nothing inside it, so its own foot is the last node there is.
func setFoot(node *yamlv3.Node, foot string) {
	if foot == "" {
		return
	}

	switch {
	case node.Kind == yamlv3.MappingNode && len(node.Content) >= 2:
		node.Content[len(node.Content)-2].FootComment = foot
	case node.Kind == yamlv3.SequenceNode && len(node.Content) > 0:
		node.Content[len(node.Content)-1].FootComment = foot
	case node.Kind == yamlv3.ScalarNode:
		node.FootComment = foot
	}
}

func valueNode(v Value) (*yamlv3.Node, error) {
	node, err := kindNode(v)
	if err != nil {
		return nil, err
	}

	node.LineComment = v.comment
	setFoot(node, v.foot)

	return node, nil
}

func kindNode(v Value) (*yamlv3.Node, error) {
	switch v.kind {
	case kindNull:
		// An empty Value, not "null": the latter would emit "key: null".
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!null", Value: ""}, nil
	case kindMap:
		return mapNode(v.doc)
	case kindSeq:
		node := &yamlv3.Node{Kind: yamlv3.SequenceNode}

		for _, item := range v.seq {
			child, err := valueNode(item)
			if err != nil {
				return nil, err
			}

			// A sequence item has no key to hang a head comment on, so it
			// goes on the item itself. That is also where a reader hands it
			// back, so the comment above a list entry survives a roundtrip.
			child.HeadComment = item.head
			node.Content = append(node.Content, child)
		}

		return node, nil
	default:
		return scalarNode(v.scalar)
	}
}

func scalarNode(scalar any) (*yamlv3.Node, error) {
	switch s := scalar.(type) {
	case string:
		node := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: s}

		if strings.Contains(s, "\n") {
			node.Style = yamlv3.LiteralStyle

			return node, nil
		}

		if needsQuoting(s) {
			node.Style = yamlv3.DoubleQuotedStyle
		}

		return node, nil
	case bool:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(s)}, nil
	default:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Value: fmt.Sprintf("%v", s)}, nil
	}
}

// plainWords are the strings a YAML 1.1 reader turns into a boolean or a null,
// held lower-cased because the comparison is case-insensitive.
var plainWords = map[string]struct{}{
	"true": {}, "false": {}, "null": {},
	"y": {}, "n": {}, "yes": {}, "no": {}, "on": {}, "off": {},
}

// needsQuoting reports whether a string would be misread without quotes,
// because it looks like a number, boolean or null, or holds a YAML special
// character.
func needsQuoting(v string) bool {
	if v == "" || v == "~" {
		return true
	}

	// YAML 1.1 reads a boolean or a null in any case, so "Yes" and "OFF" are
	// as much booleans as "yes" and "off". Matching only the lower-case forms
	// let a capitalized one out unquoted, where a reader turns it into a bool.
	if _, found := plainWords[strings.ToLower(v)]; found {
		return true
	}

	// yaml.v3 quotes a value whose ends are whitespace, but reaches for single
	// quotes to do it, and the project rule is double.
	if strings.TrimSpace(v) != v {
		return true
	}

	// If it parses as a number, it needs quoting to stay a string.

	if _, err := fmt.Sscanf(v, "%f", new(float64)); err == nil {
		// Extra check: "v" must be fully numeric (Sscanf can match a prefix).
		isNumeric := true

		for _, c := range v {
			if !((c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' || c == 'e' || c == 'E') {
				isNumeric = false
				break
			}
		}

		if isNumeric {
			return true
		}
	}

	// Characters that are special in YAML and require quoting.

	for _, c := range v {
		switch c {
		case ':', '#', '[', ']', '{', '}', ',', '&', '*', '!', '|', '>', '%', '`':
			return true
		}
	}

	// Strings starting with YAML indicators.

	if len(v) > 0 {
		switch v[0] {
		// "@" and "`" are reserved indicators: YAML does not allow a plain
		// scalar to start with either. Without them here yaml.v3 still quotes
		// the value, but picks single quotes.
		case '?', '-', '"', '\'', '@', '`':
			return true
		}
	}

	return false
}
