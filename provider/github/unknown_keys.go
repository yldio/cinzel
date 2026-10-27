// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

// droppedKey is one key the schema does not declare, with where it was
// written so the report can name it.
type droppedKey struct {
	path string
	line int
}

// stripUnknownKeys removes every key the shape does not declare, at any depth,
// and returns what it removed.
//
// The shape structs are the schema, so the accepted keys are read off them
// rather than kept in a second list that has to be edited in step. A field
// declared as "any" is opaque: whatever is under it is the author's and is
// left alone.
//
// Reported over several levels on purpose. The strict decode this stands in
// for stops at the first key it does not know, so a file with one bad key per
// job had to be run, fixed and run again for each of them.
func stripUnknownKeys(raw map[string]any, shape any, lines map[string]int) []droppedKey {
	dropped := []droppedKey{}

	stripInto(raw, reflect.TypeOf(shape), "", lines, &dropped)

	sort.Slice(dropped, func(i, j int) bool { return dropped[i].path < dropped[j].path })

	return dropped
}

func stripInto(raw map[string]any, shape reflect.Type, path string, lines map[string]int, dropped *[]droppedKey) {
	for key, value := range raw {
		at := key

		if path != "" {
			at = path + "." + key
		}

		field, known := fieldShape(shape, key)

		if !known {
			*dropped = append(*dropped, droppedKey{path: at, line: lines[at]})
			delete(raw, key)

			continue
		}

		stripValue(value, field, at, lines, dropped)
	}
}

// stripValue follows a value into whatever its field declares: a nested
// struct, a map of them, or a list of them.
func stripValue(value any, field reflect.Type, path string, lines map[string]int, dropped *[]droppedKey) {
	for field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	switch field.Kind() {
	case reflect.Struct:
		if nested, ok := toStringAnyMap(value); ok {
			stripInto(nested, field, path, lines, dropped)
		}
	case reflect.Map:
		nested, ok := toStringAnyMap(value)

		if !ok {
			return
		}

		for key, child := range nested {
			stripValue(child, field.Elem(), path+"."+key, lines, dropped)
		}
	case reflect.Slice:
		items, ok := value.([]any)

		if !ok {
			return
		}

		for i, item := range items {
			stripValue(item, field.Elem(), fmt.Sprintf("%s[%d]", path, i), lines, dropped)
		}
	}
}

// fieldShape returns the type a key's value is checked against, and whether
// the shape declares the key at all.
func fieldShape(shape reflect.Type, key string) (reflect.Type, bool) {
	for shape.Kind() == reflect.Ptr {
		shape = shape.Elem()
	}

	if shape.Kind() != reflect.Struct {
		return nil, false
	}

	for i := 0; i < shape.NumField(); i++ {
		name, _, _ := strings.Cut(shape.Field(i).Tag.Get("yaml"), ",")

		if name == key {
			return shape.Field(i).Type, true
		}
	}

	return nil, false
}

// keyLines maps the path of every key in a document to the line it was
// written on, so a dropped key can be reported where its author will find it.
//
// The decoded map holds no positions, and the comment tree keeps a line only
// for a key that carried a comment. This walks the same node tree once more
// for the rest.
func keyLines(node *yamlv3.Node) map[string]int {
	lines := map[string]int{}

	collectKeyLines(node, "", lines)

	return lines
}

func collectKeyLines(node *yamlv3.Node, path string, lines map[string]int) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yamlv3.MappingNode:
		for i := 0; i+1 < len(node.Content); i += 2 {
			key, value := node.Content[i], node.Content[i+1]
			at := key.Value

			if path != "" {
				at = path + "." + key.Value
			}

			lines[at] = key.Line

			collectKeyLines(value, at, lines)
		}
	case yamlv3.SequenceNode:
		for i, item := range node.Content {
			collectKeyLines(item, fmt.Sprintf("%s[%d]", path, i), lines)
		}
	}
}

// reportDroppedKeys names each key the conversion dropped, with the file and
// the line it was written on.
func reportDroppedKeys(site commentSite, dropped []droppedKey) {
	for _, key := range dropped {
		where := site.child(key.path).String(key.line)

		if where == "" {
			where = key.path
		}

		warnf("dropping %s: the schema does not declare it, so it cannot be written to HCL and will be gone from the YAML the next parse writes", where)
	}
}
