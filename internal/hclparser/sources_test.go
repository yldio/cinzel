// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
)

// The comment used to be read back off disk, once per attribute, so a file
// edited after the parse was read in its new shape and every lookup cost a
// syscall. It is taken from what was parsed now.
func TestTrailingComment(t *testing.T) {
	const src = "a = 1 # first\nb = 2\nc = 3 # last"

	for _, tc := range []struct {
		name string
		file string
		end  int
		want string
	}{
		{name: "a comment on the line", file: "in.hcl", end: 5, want: "# first"},
		{name: "no comment on the line", file: "in.hcl", end: 19, want: ""},
		{name: "a comment on the last line, which has no newline", file: "in.hcl", end: 25, want: "# last"},
		{name: "a file that was never parsed", file: "other.hcl", end: 5, want: ""},
		{name: "a range running past the end", file: "in.hcl", end: len(src) + 1, want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hv := NewHCLVars()
			hv.SetSources(map[string][]byte{"in.hcl": []byte(src)})

			r := hcl.Range{Filename: tc.file, End: hcl.Pos{Byte: tc.end}}

			if got := hv.TrailingComment(r); got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// Nothing set the sources, so there is nothing to read; the lookup must not
// reach for the file instead.
func TestTrailingCommentWithoutSources(t *testing.T) {
	hv := NewHCLVars()

	if got := hv.TrailingComment(hcl.Range{Filename: "in.hcl"}); got != "" {
		t.Errorf("want no comment, got %q", got)
	}
}
