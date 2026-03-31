// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"unicode/utf8"

	yamlv3 "gopkg.in/yaml.v3"
)

var pipelineKeyOrder = []string{"stages", "variables", "workflow", "default", "include"}

func marshalPipelineYAML(pipeline map[string]any) ([]byte, error) {
	root, err := pipelineMapNode(pipeline)
	if err != nil {
		return nil, err
	}

	doc := &yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{root}}

	var buf bytes.Buffer
	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)

	if err := enc.Encode(doc); err != nil {
		return nil, err
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	out := bytes.ReplaceAll(buf.Bytes(), []byte(": {}\n"), []byte(":\n"))

	return unescapeYAMLUnicode(out), nil
}

// unescapeYAMLUnicode replaces \uXXXX and \UXXXXXXXX escape sequences in YAML
// output with their raw UTF-8 equivalents for characters above U+009F.
// gopkg.in/yaml.v3 escapes supplementary-plane characters (emoji etc.) because
// its is_printable helper only handles 3-byte UTF-8 sequences. Replacing the
// escapes restores readable output without changing the YAML semantics.
func unescapeYAMLUnicode(src []byte) []byte {
	return reYAMLUnicodeEscape.ReplaceAllFunc(src, func(match []byte) []byte {
		n, err := strconv.ParseInt(string(match[2:]), 16, 32)
		if err != nil || n <= 0x9F || !utf8.ValidRune(rune(n)) {
			return match
		}

		var buf [utf8.UTFMax]byte
		l := utf8.EncodeRune(buf[:], rune(n))

		return append([]byte(nil), buf[:l]...)
	})
}

var reYAMLUnicodeEscape = regexp.MustCompile(`\\U[0-9A-Fa-f]{8}|\\u[0-9A-Fa-f]{4}`)

func pipelineMapNode(pipeline map[string]any) (*yamlv3.Node, error) {
	node := &yamlv3.Node{Kind: yamlv3.MappingNode}
	seen := map[string]struct{}{}

	for _, key := range pipelineKeyOrder {
		val, ok := pipeline[key]

		if !ok {
			continue
		}

		if err := appendMappingPair(node, key, val); err != nil {
			return nil, err
		}
		seen[key] = struct{}{}
	}

	jobs := make([]string, 0)

	for key, val := range pipeline {
		if _, ok := seen[key]; ok {
			continue
		}

		if _, ok := val.(map[string]any); ok {
			jobs = append(jobs, key)
		}
	}
	sort.Strings(jobs)

	for _, job := range jobs {
		if err := appendMappingPair(node, job, pipeline[job]); err != nil {
			return nil, err
		}
		seen[job] = struct{}{}
	}

	remaining := make([]string, 0)

	for key := range pipeline {
		if _, ok := seen[key]; ok {
			continue
		}
		remaining = append(remaining, key)
	}
	sort.Strings(remaining)

	for _, key := range remaining {
		if err := appendMappingPair(node, key, pipeline[key]); err != nil {
			return nil, err
		}
	}

	return node, nil
}

func appendMappingPair(node *yamlv3.Node, key string, value any) error {
	valueNode, err := toYAMLNode(value)
	if err != nil {
		return err
	}
	keyNode := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: key}
	node.Content = append(node.Content, keyNode, valueNode)

	return nil
}

func toYAMLNode(value any) (*yamlv3.Node, error) {
	switch v := value.(type) {
	case nil:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!null", Value: "null"}, nil
	case string:
		node := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: v}

		if stringNeedsQuoting(v) {
			node.Style = yamlv3.DoubleQuotedStyle
		}

		return node, nil
	case bool:
		if v {
			return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!bool", Value: "true"}, nil
		}

		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!bool", Value: "false"}, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Value: fmt.Sprintf("%v", v)}, nil
	case []any:
		node := &yamlv3.Node{Kind: yamlv3.SequenceNode}

		for _, item := range v {
			child, err := toYAMLNode(item)
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, child)
		}

		return node, nil
	case map[string]any:
		return genericMapNode(v)
	case map[any]any:
		m := make(map[string]any, len(v))

		for rawKey, rawValue := range v {
			key, ok := rawKey.(string)

			if !ok {
				return nil, fmt.Errorf("unsupported non-string YAML key type %T", rawKey)
			}
			m[key] = rawValue
		}

		return genericMapNode(m)
	default:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Value: fmt.Sprintf("%v", v)}, nil
	}
}

func genericMapNode(mapping map[string]any) (*yamlv3.Node, error) {
	keys := make([]string, 0, len(mapping))

	for key := range mapping {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	node := &yamlv3.Node{Kind: yamlv3.MappingNode}

	for _, key := range keys {
		if err := appendMappingPair(node, key, mapping[key]); err != nil {
			return nil, err
		}
	}

	return node, nil
}

func stringNeedsQuoting(v string) bool {
	if v == "" || v == "true" || v == "false" || v == "null" || v == "~" ||
		v == "yes" || v == "no" || v == "on" || v == "off" {
		return true
	}

	for _, c := range v {
		switch c {
		case ':', '#', '[', ']', '{', '}', ',', '&', '*', '!', '|', '>', '%', '`':
			return true
		}
	}

	if len(v) > 0 {
		switch v[0] {
		case '?', '-', '"', '\'':
			return true
		}
	}

	return false
}
