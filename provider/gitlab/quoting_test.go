// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"strings"
	"testing"

	yamlv3 "gopkg.in/yaml.v3"
)

func encodePair(t *testing.T, key string, value any) string {
	t.Helper()

	node := &yamlv3.Node{Kind: yamlv3.MappingNode}

	if err := appendMappingPair(node, key, value); err != nil {
		t.Fatal(err)
	}

	out, err := yamlv3.Marshal(node)
	if err != nil {
		t.Fatal(err)
	}

	return strings.TrimSpace(string(out))
}

// A capitalized boolean went out bare, and a value padded with spaces was left
// to yaml.v3, which reaches for single quotes.
func TestValueQuotingGaps(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "capital Yes", value: "Yes", want: `k: "Yes"`},
		{name: "capital Off", value: "Off", want: `k: "Off"`},
		{name: "single letter y", value: "y", want: `k: "y"`},
		{name: "single letter N", value: "N", want: `k: "N"`},
		{name: "leading and trailing space", value: " padded ", want: `k: " padded "`},
		{name: "an ordinary word is left alone", value: "yesterday", want: "k: yesterday"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := encodePair(t, "k", tc.value)

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}

			if strings.Contains(got, "'") {
				t.Errorf("single quotes in %q", got)
			}
		})
	}
}

// Keys never went through a quoting check, so yaml.v3 single-quoted the ones
// that needed quoting. A colon inside a key is legal, though, and GitLab job
// names rely on it, so those stay plain.
func TestKeyQuoting(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  string
		want string
	}{
		{name: "a comment character", key: "#x", want: `"#x": v`},
		{name: "a colon and a space", key: "a: b", want: `"a: b": v`},
		{name: "a trailing colon", key: "a:", want: `"a:": v`},
		{name: "padded with spaces", key: " pad ", want: `" pad ": v`},
		{name: "a capitalized boolean", key: "Yes", want: `"Yes": v`},
		{name: "a flow indicator", key: "a,b", want: `"a,b": v`},
		// A colon only ends a key when a space follows it.
		{name: "a job name with a colon", key: "test:unit", want: "test:unit: v"},
		{name: "a template name", key: ".go-base", want: ".go-base: v"},
		{name: "an ordinary name", key: "build-app", want: "build-app: v"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := encodePair(t, tc.key, "v")

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}

			if strings.Contains(got, "'") {
				t.Errorf("single quotes in %q", got)
			}
		})
	}
}
