// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

import (
	"strings"
	"testing"
)

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

// yaml.v3 quotes a reserved indicator on its own, but reaches for single
// quotes. The project writes double quotes or none.
func TestEncodeQuotesReservedIndicatorsWithDoubleQuotes(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"@daily", "k: \"@daily\"\n"},
		{"`x", "k: \"`x\"\n"},
		{"a@b.com", "k: a@b.com\n"},
	} {
		doc := New()
		doc.Set("k", Scalar(tc.in))

		got, err := Encode(doc)
		if err != nil {
			t.Fatal(err)
		}

		if string(got) != tc.want {
			t.Errorf("Encode(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestEncodeKeepsAuthoredEscape covers a string holding the four characters
// "\u00e9": the encoder doubles the backslash, and rewriting the escape
// would leave a "\é" the YAML parser rejects.
func TestEncodeKeepsAuthoredEscape(t *testing.T) {
	doc := New()
	doc.Set("run", Scalar(`echo '{"msg": "caf\u00e9"}'`))

	got, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}

	want := "run: \"echo '{\\\"msg\\\": \\\"caf\\\\u00e9\\\"}'\"\n"

	if string(got) != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

// An empty collection has no child to carry the comment that closes it. It
// used to match no case in setFoot at all, so the comment was dropped: an
// empty "permissions" block is the live way to reach that, since it is the one
// empty map the GitHub provider keeps rather than collapsing to null.
func TestFootCommentSurvivesAnEmptyCollection(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value Value
	}{
		{"empty map", Map(New(), WithFootComment("# locked down on purpose"))},
		{"empty sequence", Seq(nil, WithFootComment("# locked down on purpose"))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := New()
			d.Set("permissions", tc.value)

			out, err := Encode(d)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(out), "# locked down on purpose") {
				t.Errorf("foot comment was dropped:\n%s", out)
			}
		})
	}
}
