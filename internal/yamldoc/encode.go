// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	yamlv3 "gopkg.in/yaml.v3"
)

// Encode renders the document as YAML. Items are emitted in document order,
// inline comments are attached to their value, a null encodes as a bare "key:"
// and an empty map as "key: {}". Nothing in the output is recovered by
// rewriting encoded bytes.
func Encode(d *Doc) ([]byte, error) {
	root, err := mapNode(d)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(&yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{root}}); err != nil {
		return nil, err
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return unescapeUnicode(buf.Bytes()), nil
}

func mapNode(d *Doc) (*yamlv3.Node, error) {
	node := &yamlv3.Node{Kind: yamlv3.MappingNode}

	for _, it := range d.items {
		value, err := valueNode(it.value)
		if err != nil {
			return nil, err
		}

		key := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: it.key}
		node.Content = append(node.Content, key, value)
	}

	return node, nil
}

func valueNode(v Value) (*yamlv3.Node, error) {
	node, err := kindNode(v)
	if err != nil {
		return nil, err
	}

	node.LineComment = v.comment

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

// needsQuoting reports whether a string would be misread without quotes,
// because it looks like a number, boolean or null, or holds a YAML special
// character.
func needsQuoting(v string) bool {
	if v == "" || v == "true" || v == "false" || v == "null" || v == "~" ||
		v == "yes" || v == "no" || v == "on" || v == "off" {
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
		case '?', '-', '"', '\'':
			return true
		}
	}

	return false
}

// unescapeUnicode replaces \uXXXX and \UXXXXXXXX escapes with their raw UTF-8
// equivalents for characters above U+009F. gopkg.in/yaml.v3 escapes
// supplementary-plane characters (emoji etc.) because its is_printable helper
// only handles 3-byte UTF-8 sequences. Replacing the escapes restores readable
// output without changing the YAML semantics.
func unescapeUnicode(src []byte) []byte {
	return reUnicodeEscape.ReplaceAllFunc(src, func(match []byte) []byte {
		n, err := strconv.ParseInt(string(match[2:]), 16, 32)
		if err != nil || n <= 0x9F || !utf8.ValidRune(rune(n)) {
			return match
		}

		var buf [utf8.UTFMax]byte
		l := utf8.EncodeRune(buf[:], rune(n))

		return append([]byte(nil), buf[:l]...)
	})
}

var reUnicodeEscape = regexp.MustCompile(`\\U[0-9A-Fa-f]{8}|\\u[0-9A-Fa-f]{4}`)
