// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// A comment is not part of any decoded value, so the source is lexed to find
// one. The collector used to live in the github provider and serve job block
// headers alone; every other comment in a file went unread.
func TestHeadComment(t *testing.T) {
	const src = `# above the block
# second line
job "build" {
  # above the attribute
  run = "make"

  # after a blank line
  shell = "bash" # trailing, not a head comment
  other = 1
}
`

	for _, tc := range []struct {
		name string
		file string
		line int
		want string
	}{
		{name: "the run above a block", file: "in.hcl", line: 3, want: "# above the block\n# second line"},
		{name: "one line above an attribute", file: "in.hcl", line: 5, want: "# above the attribute"},
		{name: "a blank line does not break a run reaching it", file: "in.hcl", line: 8, want: "# after a blank line"},
		{name: "a blank line above leaves nothing", file: "in.hcl", line: 7, want: ""},
		{name: "a comment trailing code is not a head comment", file: "in.hcl", line: 9, want: ""},
		{name: "nothing above the first line", file: "in.hcl", line: 1, want: ""},
		{name: "a file that was never parsed", file: "other.hcl", line: 3, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hv := NewHCLVars()
			hv.SetSources(map[string][]byte{"in.hcl": []byte(src)})

			r := hcl.Range{Filename: tc.file, Start: hcl.Pos{Line: tc.line}}

			if got := hv.HeadComment(r); got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// The text is the author's. Whatever spacing or hash count they wrote comes
// back as they wrote it.
func TestHeadCommentIsVerbatim(t *testing.T) {
	const src = "#no space\n##  banner   style\njob \"build\" {\n}\n"

	hv := NewHCLVars()
	hv.SetSources(map[string][]byte{"in.hcl": []byte(src)})

	want := "#no space\n##  banner   style"
	got := hv.HeadComment(hcl.Range{Filename: "in.hcl", Start: hcl.Pos{Line: 3}})

	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

// A file is lexed once and the tokens held, so a body with forty attributes
// does not lex it forty times.
func TestHeadCommentLexesEachFileOnce(t *testing.T) {
	hv := NewHCLVars()
	hv.SetSources(map[string][]byte{"in.hcl": []byte("# a\nx = 1\n")})

	hv.HeadComment(hcl.Range{Filename: "in.hcl", Start: hcl.Pos{Line: 2}})
	hv.HeadComment(hcl.Range{Filename: "in.hcl", Start: hcl.Pos{Line: 2}})

	if len(hv.lexed) != 1 {
		t.Fatalf("want one file lexed, got %d", len(hv.lexed))
	}
}

// A file that does not lex has no comments to give. The decode reports the
// syntax error with the detail this pass deliberately does not collect, so
// nothing is said here and nothing crashes.
func TestHeadCommentOnUnlexableSource(t *testing.T) {
	hv := NewHCLVars()
	hv.SetSources(map[string][]byte{"in.hcl": []byte("# a\nx = \"unterminated\n")})

	if got := hv.HeadComment(hcl.Range{Filename: "in.hcl", Start: hcl.Pos{Line: 2}}); got != "" {
		t.Errorf("want no comment, got %q", got)
	}
}
