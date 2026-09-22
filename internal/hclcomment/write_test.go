// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclcomment

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func TestLineAddsAHashOnlyWhereOneIsMissing(t *testing.T) {
	cases := []struct {
		name string
		line string
		want string
	}{
		{name: "plain prose", line: "why this is here", want: "# why this is here"},
		{name: "already a comment", line: "# why this is here", want: "# why this is here"},
		{name: "indented comment", line: "  # indented", want: "  # indented"},
		{name: "two hashes", line: "## loud", want: "## loud"},
		{name: "a hash further along", line: "runner #1", want: "# runner #1"},
		{name: "empty", line: "", want: "# "},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := Line(tt.line); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestWriteLeadingWritesEveryLineAsAComment(t *testing.T) {
	f := hclwrite.NewEmptyFile()

	WriteLeading(f.Body(), "first\n# second\nthird")
	f.Body().SetAttributeValue("k", cty.StringVal("v"))

	want := "# first\n# second\n# third\nk = \"v\"\n"

	if got := string(f.Bytes()); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	mustParse(t, f.Bytes())
}

func TestWriteLeadingWritesNothingForAnEmptyComment(t *testing.T) {
	f := hclwrite.NewEmptyFile()

	WriteLeading(f.Body(), "")
	f.Body().SetAttributeValue("k", cty.StringVal("v"))

	if got := string(f.Bytes()); got != "k = \"v\"\n" {
		t.Fatalf("expected the attribute alone, got %q", got)
	}
}

func TestTrailingLandsOnTheSameLineAsTheValue(t *testing.T) {
	f := hclwrite.NewEmptyFile()

	f.Body().SetAttributeRaw("k", Trailing(hclwrite.TokensForValue(cty.StringVal("v")), "why"))

	// The spacing between the value and the "#" is hclwrite's to choose, so
	// what is checked is that the comment stayed on the attribute's line.
	got := strings.TrimRight(string(f.Bytes()), "\n")

	if strings.Contains(got, "\n") {
		t.Fatalf("expected one line, got %q", got)
	}

	if !strings.HasPrefix(got, "k = \"v\"") || !strings.HasSuffix(got, "# why") {
		t.Fatalf("expected the value then the comment, got %q", got)
	}

	mustParse(t, f.Bytes())
}

func TestTrailingLeavesTheTokensAloneForAnEmptyComment(t *testing.T) {
	tokens := hclwrite.TokensForValue(cty.StringVal("v"))

	if got := Trailing(tokens, ""); len(got) != len(tokens) {
		t.Fatalf("expected %d tokens, got %d", len(tokens), len(got))
	}
}

// TestTrailingTakesOneLine pins the contract the callers rely on rather than
// the behaviour a multi-line comment would get.
//
// Trailing writes its text as one run of bytes, so a newline inside it puts the
// rest on a line of its own with no "#" to make it a comment, and the generated
// file no longer parses. Nothing reaches it that way: the only caller reads a
// YAML line comment, and a line comment ends at its line. The test says so, so
// that a caller that one day passes several lines finds this note rather than
// an unreadable file.
func TestTrailingTakesOneLine(t *testing.T) {
	f := hclwrite.NewEmptyFile()

	f.Body().SetAttributeRaw("k", Trailing(hclwrite.TokensForValue(cty.StringVal("v")), "one\ntwo"))

	_, diags := hclsyntax.ParseConfig(f.Bytes(), "p.hcl", hcl.InitialPos)

	if !diags.HasErrors() {
		t.Fatalf("a multi-line trailing comment now parses: either Trailing splits its lines, or this test is stale:\n%s", f.Bytes())
	}

	if !strings.Contains(string(f.Bytes()), "\ntwo") {
		t.Fatalf("expected the second line to land uncommented, got %q", f.Bytes())
	}
}

func mustParse(t *testing.T, src []byte) {
	t.Helper()

	if _, diags := hclsyntax.ParseConfig(src, "p.hcl", hcl.InitialPos); diags.HasErrors() {
		t.Fatalf("generated HCL does not parse: %s\n%s", diags.Error(), src)
	}
}
