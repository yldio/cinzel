// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A comment is not a reference. Counting every "version =" in the text left
// more versions than actions, and the whole file was refused.
func TestCommentMentioningAVersionIsIgnored(t *testing.T) {
	for _, comment := range []string{
		`// TODO: was on version = "v3"`,
		`# was on version = "v3"`,
		`/* version = "v3" */`,
	} {
		t.Run(comment, func(t *testing.T) {
			content := `step "a" {
  ` + comment + `
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

			refs, err := findActionRefs(content)
			if err != nil {
				t.Fatalf("a comment should not fail the parse: %v", err)
			}

			if len(refs) != 1 {
				t.Fatalf("expected one ref, got %d: %+v", len(refs), refs)
			}

			if refs[0].Action != "actions/checkout" || refs[0].Version != "v4" {
				t.Errorf("got %s@%s", refs[0].Action, refs[0].Version)
			}
		})
	}
}

// The file is refused as a whole when the parse fails, so a comment left the
// action on a moving tag. PinDirectory reports that as a warning and exits 0,
// which is how it goes unnoticed.
func TestAnActionBesideACommentIsStillPinned(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "steps.hcl")

	content := `step "checkout" {
  // TODO: was on version = "v3" before the bump
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &mockResolver{shas: map[string]string{
		"actions/checkout@v4": "1111111111111111111111111111111111111111",
	}}

	var buf bytes.Buffer

	results, err := PinFile(context.Background(), path, resolver, &buf, false)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("expected one action to be pinned, got %d", len(results))
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(after), `version = "1111111111111111111111111111111111111111" # v4`) {
		t.Errorf("the action was not pinned:\n%s", after)
	}

	// The comment is left as it was written.
	if !strings.Contains(string(after), `// TODO: was on version = "v3" before the bump`) {
		t.Errorf("the comment was disturbed:\n%s", after)
	}
}

// A uses block holding only one half of the pair names no action to pin, and
// must not consume the next block's version.
func TestIncompleteUsesBlockIsSkipped(t *testing.T) {
	content := `step "a" {
  uses {
    action = "acme/one"
  }
}

step "b" {
  uses {
    action  = "acme/two"
    version = "v2"
  }
}
`

	refs, err := findActionRefs(content)
	if err != nil {
		t.Fatal(err)
	}

	if len(refs) != 1 {
		t.Fatalf("expected one complete ref, got %d: %+v", len(refs), refs)
	}

	if refs[0].Action != "acme/two" || refs[0].Version != "v2" {
		t.Errorf("got %s@%s, the halves were paired across blocks", refs[0].Action, refs[0].Version)
	}
}

// Refs are returned in the order they appear, which is what the rewrite
// relies on when it splices edits back to front.
//
// This is a regression guard, not a gate: hclsyntax keeps Blocks in source
// order and the walk follows it, so the sort in findActionRefs is not what
// makes this pass. Removing the sort leaves it passing in every shape tried,
// nesting and sibling order included. The sort stays because the rewrite
// depends on the ordering and should say so, rather than resting on a
// property of hclsyntax that nothing here states.
func TestRefsComeBackInSourceOrder(t *testing.T) {
	content := `step "a" {
  uses {
    action  = "acme/one"
    version = "v1"
  }
}

step "b" {
  uses {
    version = "v2"
    action  = "acme/two"
  }
}
`

	refs, err := findActionRefs(content)
	if err != nil {
		t.Fatal(err)
	}

	if len(refs) != 2 {
		t.Fatalf("expected two refs, got %d", len(refs))
	}

	if refs[0].Action != "acme/one" || refs[1].Action != "acme/two" {
		t.Errorf("out of order: %s then %s", refs[0].Action, refs[1].Action)
	}

	if refs[0].start >= refs[1].start {
		t.Errorf("offsets not ascending: %d then %d", refs[0].start, refs[1].start)
	}
}
