// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"bytes"
	"fmt"
	"sort"

	yamlv3 "gopkg.in/yaml.v3"
)

var workflowKeyOrder = []string{
	"name",
	"run-name",
	"on",
	"permissions",
	"env",
	"defaults",
	"concurrency",
	"jobs",
}

func marshalWorkflowYAML(workflow map[string]any) ([]byte, error) {
	rootNode, err := workflowMapNode(workflow)
	if err != nil {
		return nil, err
	}

	doc := &yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{rootNode}}

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
	return out, nil
}

func workflowMapNode(workflow map[string]any) (*yamlv3.Node, error) {
	node := &yamlv3.Node{Kind: yamlv3.MappingNode}

	seen := map[string]struct{}{}
	for _, key := range workflowKeyOrder {
		value, ok := workflow[key]
		if !ok {
			continue
		}

		if err := appendMappingPair(node, key, value); err != nil {
			return nil, err
		}

		seen[key] = struct{}{}
	}

	remaining := make([]string, 0, len(workflow))
	for key := range workflow {
		if _, ok := seen[key]; ok {
			continue
		}

		remaining = append(remaining, key)
	}

	sort.Strings(remaining)
	for _, key := range remaining {
		if err := appendMappingPair(node, key, workflow[key]); err != nil {
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
		stringMap := make(map[string]any, len(v))
		for rawKey, rawValue := range v {
			key, ok := rawKey.(string)
			if !ok {
				return nil, fmt.Errorf("unsupported non-string YAML key type %T", rawKey)
			}

			stringMap[key] = rawValue
		}

		return genericMapNode(stringMap)
	default:
		return &yamlv3.Node{Kind: yamlv3.ScalarNode, Value: fmt.Sprintf("%v", v)}, nil
	}
}

// stringNeedsQuoting returns true if a YAML string value would be
// misinterpreted without quotes (e.g., looks like a number, boolean,
// null, or contains special characters).
func stringNeedsQuoting(v string) bool {
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
