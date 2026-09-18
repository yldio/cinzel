// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import "testing"

// relabel used to delete from and insert into the same map it was ranging
// over, so whether a rename whose target is another block's label was visited
// again was undefined: the same input moved a comment onto a different key
// from one run to the next.
func TestRelabelIsDeterministic(t *testing.T) {
	// A → B and B → C chain, which is the shape that made the order matter.
	keys := map[string]string{"A": "B", "B": "C"}

	for range 200 {
		mapping := &comments{children: map[string]*comments{
			"A": {own: map[string]nodeComment{"marker": {head: "// from A"}}},
			"B": {own: map[string]nodeComment{"marker": {head: "// from B"}}},
		}}

		relabel(mapping, keys)

		if len(mapping.children) != 2 {
			t.Fatalf("expected 2 children, got %d: %v", len(mapping.children), childKeys(mapping))
		}

		for label, want := range map[string]string{"B": "// from A", "C": "// from B"} {
			child, ok := mapping.children[label]

			if !ok {
				t.Fatalf("missing %q, got %v", label, childKeys(mapping))
			}

			if got := child.own["marker"].head; got != want {
				t.Fatalf("%q carries %q, want %q", label, got, want)
			}
		}
	}
}

func childKeys(mapping *comments) []string {
	out := make([]string, 0, len(mapping.children))

	for label := range mapping.children {
		out = append(out, label)
	}

	return out
}
