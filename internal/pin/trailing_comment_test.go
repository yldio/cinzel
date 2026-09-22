// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

// A "/* */" ends where it closes, so what follows it on the same line is still
// part of the version line's trailing comment. Stopping at the close left that
// second comment standing beside the one the rewrite writes, which is the
// stacked "# v5 # v4" naming a version the SHA is not.
func TestASecondCommentAfterABlockCommentIsAlsoRead(t *testing.T) {
	for _, tc := range []struct {
		name    string
		comment string
		note    string
	}{
		{name: "hash after a block comment", comment: `/* pinned */ # v4`, note: "pinned"},
		{name: "slash after a block comment", comment: `/* pinned */ // v4`, note: "pinned"},
		{name: "two block comments", comment: `/* pinned */ /* by hand */`, note: "pinned by hand"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeHCL(t, `step "a" {
  uses {
    action  = "acme/stale"
    version = "cccccccccccccccccccccccccccccccccccccccc" `+tc.comment+`
  }
}
`)

			resolver := &upgraderStub{
				latest: map[string]string{"acme/stale": "v5"},
				shas:   map[string]string{"acme/stale@v5": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},
			}

			var buf bytes.Buffer

			if _, err := UpgradeFile(context.Background(), path, resolver, &buf, false); err != nil {
				t.Fatal(err)
			}

			out, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			got := string(out)

			if _, err := findActionRefs(got); err != nil {
				t.Fatalf("the upgraded file no longer parses: %v\n%s", err, got)
			}

			if strings.Contains(got, "v4") {
				t.Errorf("the superseded tag v4 is still named in:\n%s", got)
			}

			if n := strings.Count(got, "#"); n != 1 {
				t.Errorf("expected one comment, found %d in:\n%s", n, got)
			}

			// The whole line: a second comment read but not carried is the
			// author's note deleted, and one read but not consumed is the
			// stacked comment this test is named for. Only the exact text
			// tells those two apart from the right answer.
			want := `    version = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" # v5 ` + tc.note + "\n"

			if !strings.Contains(got, want) {
				t.Errorf("version line = not %q, in:\n%s", want, got)
			}
		})
	}
}

// A "\r" is part of the line ending, not of the comment. Taken with it, the
// rewritten line came out LF in a file that is CRLF throughout, so a pin on a
// Windows checkout produced mixed endings and a diff on lines nobody edited.
func TestPinKeepsCRLFLineEndings(t *testing.T) {
	for _, tc := range []struct {
		name string
		line string
		note string
	}{
		{name: "no comment", line: "    version = \"v4\"\r\n"},
		{name: "hash comment", line: "    version = \"v4\" # note\r\n", note: " note"},
		{name: "block comment", line: "    version = \"v4\" /* note */\r\n", note: " note"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src := "step \"a\" {\r\n  uses {\r\n    action  = \"actions/checkout\"\r\n" +
				tc.line + "  }\r\n}\r\n"

			path := writeHCL(t, src)

			resolver := &mockResolver{shas: map[string]string{
				"actions/checkout@v4": "ffffffffffffffffffffffffffffffffffffffff",
			}}

			var buf bytes.Buffer

			if _, err := PinFile(context.Background(), path, resolver, &buf, false); err != nil {
				t.Fatal(err)
			}

			out, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			got := string(out)

			if lone := strings.Count(got, "\n") - strings.Count(got, "\r\n"); lone != 0 {
				t.Errorf("%d line(s) came out LF in a CRLF file:\n%q", lone, got)
			}

			// The whole line, not a prefix of it: taking one character too
			// many off the old comment leaves its tail standing after the new
			// one, which a Contains check on the prefix does not see.
			want := "    version = \"ffffffffffffffffffffffffffffffffffffffff\" # v4" + tc.note + "\r\n"

			if !strings.Contains(got, want) {
				t.Errorf("version line = not %q, in:\n%q", want, got)
			}
		})
	}
}
