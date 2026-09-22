// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package pin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// upgraderStub answers both halves of an upgrade from fixed tables, so a
// single action can be made to fail or to sit already on the latest tag
// while the others around it resolve.
type upgraderStub struct {
	latest map[string]string
	shas   map[string]string
}

func (m *upgraderStub) ResolveTag(_ context.Context, owner, repo, tag string) (string, error) {
	if sha, ok := m.shas[fmt.Sprintf("%s/%s@%s", owner, repo, tag)]; ok {
		return sha, nil
	}

	return "", fmt.Errorf("tag not found: %s/%s@%s", owner, repo, tag)
}

func (m *upgraderStub) LatestTag(_ context.Context, owner, repo string) (string, error) {
	if tag, ok := m.latest[owner+"/"+repo]; ok {
		return tag, nil
	}

	return "", fmt.Errorf("no releases for %s/%s", owner, repo)
}

func writeHCL(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "steps.hcl")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	return path
}

// versionFor returns the version assigned to action in content, so a test can
// say which line moved rather than only that some line did.
func versionFor(t *testing.T, content, action string) string {
	t.Helper()

	idx := strings.Index(content, fmt.Sprintf("%q", action))
	if idx < 0 {
		t.Fatalf("action %s not found in:\n%s", action, content)
	}

	rest := content[idx:]

	vIdx := strings.Index(rest, "version = ")
	if vIdx < 0 {
		t.Fatalf("no version after action %s in:\n%s", action, content)
	}

	line := rest[vIdx:]
	if nl := strings.Index(line, "\n"); nl >= 0 {
		line = line[:nl]
	}

	return strings.TrimSpace(strings.TrimPrefix(line, "version = "))
}

// A pin that cannot resolve one action must not move that action's line. The
// SHA belongs to the action it was resolved for, and writing it anywhere else
// pins a step to a commit from a different repository.
func TestFailedPinLeavesItsOwnLineAlone(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "acme/private"
    version = "v4"
  }
}

step "b" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}
`)

	// acme/private is absent from the table, so it fails to resolve.
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

	if v := versionFor(t, got, "acme/private"); v != `"v4"` {
		t.Errorf("acme/private failed to resolve but its version became %s", v)
	}

	if v := versionFor(t, got, "actions/checkout"); !strings.HasPrefix(v, `"ffffffffffffffffffffffffffffffffffffffff"`) {
		t.Errorf("actions/checkout resolved but its version is %s", v)
	}
}

// An upgrade skips an action already on the latest tag. The action that does
// need moving must still be the one that moves.
func TestUpgradeMovesTheActionItResolved(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "acme/current"
    version = "v4"
  }
}

step "b" {
  uses {
    action  = "acme/stale"
    version = "v4"
  }
}
`)

	resolver := &upgraderStub{
		latest: map[string]string{"acme/current": "v4", "acme/stale": "v5"},
		shas: map[string]string{
			"acme/current@v4": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"acme/stale@v5":   "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		},
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

	if v := versionFor(t, got, "acme/current"); v != `"v4"` {
		t.Errorf("acme/current was already on the latest tag but its version became %s", v)
	}

	if v := versionFor(t, got, "acme/stale"); !strings.HasPrefix(v, `"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"`) {
		t.Errorf("acme/stale should have moved to v5 but its version is %s", v)
	}
}

// Upgrading twice must leave one comment naming the version actually pinned.
// A stacked "# v5 # v4" reads as a claim about the SHA that is not true.
func TestUpgradingTwiceLeavesOneComment(t *testing.T) {
	path := writeHCL(t, `step "b" {
  uses {
    action  = "acme/stale"
    version = "cccccccccccccccccccccccccccccccccccccccc" # v4
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

	if n := strings.Count(got, "#"); n != 1 {
		t.Errorf("expected one comment, found %d in:\n%s", n, got)
	}

	if strings.Contains(got, "v4") {
		t.Errorf("the superseded tag v4 is still named in:\n%s", got)
	}

	if !strings.Contains(got, `version = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" # v5`) {
		t.Errorf("expected the line to name v5 alone, got:\n%s", got)
	}
}

// The ordinary case has to keep working: every action resolves, every line
// moves to its own SHA, and the tag is recorded beside it.
func TestEveryResolvedActionGetsItsOwnSHA(t *testing.T) {
	path := writeHCL(t, `step "a" {
  uses {
    action  = "actions/checkout"
    version = "v4"
  }
}

step "b" {
  uses {
    action  = "actions/setup-go"
    version = "v4"
  }
}
`)

	resolver := &mockResolver{shas: map[string]string{
		"actions/checkout@v4": "1111111111111111111111111111111111111111",
		"actions/setup-go@v4": "2222222222222222222222222222222222222222",
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

	if v := versionFor(t, got, "actions/checkout"); v != `"1111111111111111111111111111111111111111" # v4` {
		t.Errorf("actions/checkout got %s", v)
	}

	if v := versionFor(t, got, "actions/setup-go"); v != `"2222222222222222222222222222222222222222" # v4` {
		t.Errorf("actions/setup-go got %s", v)
	}
}

// A block comment on the version line is replaced along with the rest of it.
// The replacement carries its own "# tag" comment, and leaving a "/*" standing
// after it commented out the opening of the block comment while its "*/" stayed
// on a line of its own, so the file no longer parsed. The pin itself reported
// success, and the breakage surfaced on the next parse.
func TestBlockCommentOnTheVersionLineIsReplaced(t *testing.T) {
	for _, tc := range []struct {
		name    string
		comment string
	}{
		{name: "on one line", comment: `/* pinned by hand */`},
		{name: "over two lines", comment: "/* pinned\n    by hand */"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeHCL(t, `step "a" {
  uses {
    action  = "actions/checkout"
    version = "v4" `+tc.comment+`
  }
}
`)

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

			if _, err := findActionRefs(string(out)); err != nil {
				t.Fatalf("the pinned file no longer parses: %v\n%s", err, out)
			}

			if strings.Contains(string(out), "by hand") {
				t.Errorf("the old comment was left beside the new one:\n%s", out)
			}
		})
	}
}
