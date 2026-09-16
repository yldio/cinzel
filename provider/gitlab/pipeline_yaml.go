// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/yldio/cinzel/internal/unescape"
	yamlv3 "gopkg.in/yaml.v3"
)

var pipelineKeyOrder = []string{
	"stages", "variables", "workflow", "default", "include",
	// The five GitLab still reads outside a "default" block. They are listed
	// here so a global "cache", which is a mapping, is not mistaken for a job.
	"image", "before_script", "after_script", "cache", "services",
}

func marshalPipelineYAML(pipeline map[string]any) ([]byte, error) {
	var docs []*yamlv3.Node

	// A "spec" header has to be a document of its own, ahead of the rest of
	// the configuration. GitLab reads it nowhere else.
	if spec, hasSpec := pipeline[specKey]; hasSpec {
		rest := make(map[string]any, len(pipeline)-1)

		for key, value := range pipeline {
			if key != specKey {
				rest[key] = value
			}
		}
		pipeline = rest

		specNode, err := toYAMLNode(map[string]any{specKey: spec})
		if err != nil {
			return nil, err
		}

		docs = append(docs, &yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{specNode}})
	}

	root, err := pipelineMapNode(pipeline)
	if err != nil {
		return nil, err
	}

	docs = append(docs, &yamlv3.Node{Kind: yamlv3.DocumentNode, Content: []*yamlv3.Node{root}})

	var buf bytes.Buffer
	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)

	for _, doc := range docs {
		if err := enc.Encode(doc); err != nil {
			return nil, err
		}
	}

	if err := enc.Close(); err != nil {
		return nil, err
	}

	return unescape.Unicode(buf.Bytes()), nil
}

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

	// Keys never went through a quoting check at all, so yaml.v3 single-quoted
	// the ones that needed quoting, against the project rule.
	if keyNeedsQuoting(key) {
		keyNode.Style = yamlv3.DoubleQuotedStyle
	}

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

// keyNeedsQuoting reports whether a mapping key would be misread unquoted.
//
// A key is looser than a value: a colon inside one only ends it when a space
// follows, and GitLab job names such as "test:unit" rely on that, so running
// keys through stringNeedsQuoting would quote a large share of real pipelines.
// What does need quoting is a key that ends in a colon, holds ": ", starts a
// comment, or is one of the scalars a reader turns into a bool or a null.
func keyNeedsQuoting(key string) bool {
	if key == "" || key == "~" {
		return true
	}

	if _, found := plainWords[strings.ToLower(key)]; found {
		return true
	}

	if strings.TrimSpace(key) != key {
		return true
	}

	if strings.Contains(key, ": ") || strings.HasSuffix(key, ":") ||
		strings.Contains(key, " #") || strings.HasPrefix(key, "#") {
		return true
	}

	for _, c := range key {
		switch c {
		case '[', ']', '{', '}', ',', '&', '*', '!', '|', '>', '%', '`':
			return true
		}
	}

	switch key[0] {
	case '?', '-', '"', '\'', '@':
		return true
	}

	return false
}

// plainWords are the strings a YAML 1.1 reader turns into a boolean or a null,
// held lower-cased because the comparison is case-insensitive.
var plainWords = map[string]struct{}{
	"true": {}, "false": {}, "null": {},
	"y": {}, "n": {}, "yes": {}, "no": {}, "on": {}, "off": {},
}

func stringNeedsQuoting(v string) bool {
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

	for _, c := range v {
		switch c {
		case ':', '#', '[', ']', '{', '}', ',', '&', '*', '!', '|', '>', '%', '`':
			return true
		}
	}

	if len(v) > 0 {
		switch v[0] {
		// "@" is a reserved indicator: YAML does not allow a plain scalar to
		// start with one. Without it here yaml.v3 still quotes the value, but
		// picks single quotes.
		case '?', '-', '"', '\'', '@':
			return true
		}
	}

	return false
}
