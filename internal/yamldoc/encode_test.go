// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

import "testing"

func TestEncode(t *testing.T) {
	inner := New()
	inner.Set("contents", Scalar("read", WithComment("# only needed for private repos")))

	doc := New()
	doc.Set("name", Scalar("Fidelity ✅"))
	doc.Set("permissions", Map(New()))
	doc.Set("defaults", Null())
	doc.Set("jobs", Map(inner))
	doc.Set("script", Scalar("cat cfg <<'YAML'\noverrides: {}\nYAML\n"))
	doc.Set("list", Seq([]Value{Scalar("a"), Scalar(true), Scalar(2)}))

	// Setting an existing key replaces in place rather than reordering.
	doc.Set("name", Scalar("Fidelity ✅"))

	got, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}

	want := `name: Fidelity ✅
permissions: {}
defaults:
jobs:
  contents: read # only needed for private repos
script: |
  cat cfg <<'YAML'
  overrides: {}
  YAML
list:
  - a
  - true
  - 2
`

	if string(got) != want {
		t.Fatalf("encode mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
