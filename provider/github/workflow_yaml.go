// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"sort"

	"github.com/yldio/cinzel/internal/yamldoc"
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
	return yamldoc.Encode(workflowDoc(workflow))
}

// workflowDoc converts a workflow map into an ordered document: root keys
// follow workflowKeyOrder with the rest sorted, jobs follow the "jobsOrder"
// sentinel, and nested maps are sorted.
func workflowDoc(workflow map[string]any) *yamldoc.Doc {
	doc := yamldoc.New()
	seen := map[string]struct{}{"jobsOrder": {}}

	// "jobsOrder" is a private sentinel that must never appear in the output.
	jobOrder, _ := workflow["jobsOrder"].([]string)

	for _, key := range workflowKeyOrder {
		value, ok := workflow[key]

		if !ok {
			continue
		}

		seen[key] = struct{}{}

		if jobs, ok := value.(map[string]any); ok && key == "jobs" && len(jobOrder) > 0 {
			doc.Set(key, yamldoc.Map(orderedDoc(jobs, jobOrder)))

			continue
		}

		doc.Set(key, docValue(key, value))
	}

	appendSorted(doc, workflow, seen)

	return doc
}

// orderedDoc converts mapping with the keys in order first and the rest sorted.
func orderedDoc(mapping map[string]any, order []string) *yamldoc.Doc {
	doc := yamldoc.New()
	seen := make(map[string]struct{}, len(order))

	for _, key := range order {
		value, ok := mapping[key]

		if !ok {
			continue
		}

		doc.Set(key, docValue(key, value))
		seen[key] = struct{}{}
	}

	appendSorted(doc, mapping, seen)

	return doc
}

func appendSorted(doc *yamldoc.Doc, mapping map[string]any, seen map[string]struct{}) {
	remaining := make([]string, 0, len(mapping)-len(seen))

	for key := range mapping {
		if _, ok := seen[key]; !ok {
			remaining = append(remaining, key)
		}
	}

	sort.Strings(remaining)

	for _, key := range remaining {
		doc.Set(key, docValue(key, mapping[key]))
	}
}

// docValue converts one value. key is the key the value was found under, which
// decides whether an empty map stays explicit: an empty "permissions" does,
// because an absent or null permissions field makes GitHub Actions inherit the
// default token permissions rather than denying all access. Every other empty
// map collapses to a bare "key:".
func docValue(key string, value any, opts ...yamldoc.Opt) yamldoc.Value {
	switch v := value.(type) {
	case annotated:
		return docValue(key, v.value, append(opts, yamldoc.WithComment(v.comment))...)
	case nil:
		return yamldoc.Null(opts...)
	case map[string]any:
		if len(v) == 0 && key != "permissions" {
			return yamldoc.Null(opts...)
		}

		return yamldoc.Map(orderedDoc(v, nil), opts...)
	case map[any]any:
		stringMap := make(map[string]any, len(v))

		for rawKey, rawValue := range v {
			name, ok := rawKey.(string)

			if !ok {
				name = fmt.Sprintf("%v", rawKey)
			}

			stringMap[name] = rawValue
		}

		return docValue(key, stringMap, opts...)
	case []any:
		items := make([]yamldoc.Value, 0, len(v))

		for _, item := range v {
			items = append(items, docValue(key, item))
		}

		return yamldoc.Seq(items, opts...)
	default:
		return yamldoc.Scalar(v, opts...)
	}
}

// annotated wraps a value with an optional inline YAML comment.
// It is used to thread HCL trailing # comments through the map[string]any
// pipeline so that docValue can attach them to the document value.
type annotated struct {
	value   any
	comment string
}

// unwrapAnnotated returns the inner value if v is annotated, otherwise v itself.
// Use this before type-asserting values that may have been produced by parseBodyMap.
func unwrapAnnotated(v any) any {
	if a, ok := v.(annotated); ok {
		return a.value
	}

	return v
}

// unwrapAnnotatedMap returns a copy of m with all annotated values replaced by
// their inner values. Use before passing maps to validators that do not know
// about the annotated wrapper.
func unwrapAnnotatedMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))

	for k, v := range m {
		out[k] = unwrapAnnotated(v)
	}

	return out
}
