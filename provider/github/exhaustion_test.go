// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"strings"
	"testing"
)

// aliasBomb builds a billion-laughs document: each anchor names the one
// below it fan times, so the expanded size is fan^levels entries from an
// input of a few hundred bytes.
func aliasBomb(levels, fan int) string {
	var b strings.Builder

	b.WriteString("name: bomb\non: push\na0: &a0 [x,x,x,x,x,x,x,x,x,x]\n")

	for i := 1; i <= levels; i++ {
		fmt.Fprintf(&b, "a%d: &a%d [", i, i)

		for j := range fan {
			if j > 0 {
				b.WriteString(",")
			}

			fmt.Fprintf(&b, "*a%d", i-1)
		}

		b.WriteString("]\n")
	}

	b.WriteString("jobs:\n  build:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n")

	return b.String()
}

// TestAliasBombIsRejected pins yaml.v3's alias-expansion cap. Without it a
// few hundred bytes expand without bound before any check of ours runs.
func TestAliasBombIsRejected(t *testing.T) {
	t.Parallel()

	err := unparseErr(t, aliasBomb(11, 10))
	if err == nil {
		t.Fatal("expected the alias bomb to be rejected, got no error")
	}

	if !strings.Contains(err.Error(), "excessive aliasing") {
		t.Fatalf("expected an excessive-aliasing error, got: %v", err)
	}
}

// TestDeepNestingIsRejected pins yaml.v3's depth cap, which stops a
// recursive walk of the node tree from exhausting the stack.
func TestDeepNestingIsRejected(t *testing.T) {
	t.Parallel()

	deep := "name: deep\non: push\njobs: " +
		strings.Repeat("[", 100000) + strings.Repeat("]", 100000) + "\n"

	err := unparseErr(t, deep)
	if err == nil {
		t.Fatal("expected the deeply nested document to be rejected, got no error")
	}

	if !strings.Contains(err.Error(), "exceeded max depth") {
		t.Fatalf("expected a max-depth error, got: %v", err)
	}
}

// TestModestAliasingStillWorks keeps the two caps above from being read as
// a ban on anchors. A workflow that shares one block across jobs is normal
// and has to keep working.
func TestModestAliasingStillWorks(t *testing.T) {
	t.Parallel()

	hcl := unparse(t, "name: shared\non: push\njobs:\n"+
		"  first:\n    runs-on: &runner ubuntu-latest\n"+
		"    steps:\n      - run: echo one\n"+
		"  second:\n    runs-on: *runner\n"+
		"    steps:\n      - run: echo two\n")

	if count := strings.Count(hcl, "ubuntu-latest"); count != 2 {
		t.Fatalf("expected the anchor to resolve for both jobs, got %d occurrences in:\n%s", count, hcl)
	}
}
