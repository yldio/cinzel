// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// aliasBomb builds a billion-laughs pipeline: each anchor names the one
// below it fan times, so the expanded size is fan^levels entries from an
// input of a few hundred bytes.
func aliasBomb(levels, fan int) string {
	var b strings.Builder

	b.WriteString("stages: [build]\na0: &a0 [x,x,x,x,x,x,x,x,x,x]\n")

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

	fmt.Fprintf(&b, "job:\n  stage: build\n  script: *a%d\n", levels)

	return b.String()
}

// unparseGitLab writes the YAML to a file, unparses it and returns the HCL
// together with the error, if any.
func unparseGitLab(t *testing.T, yaml string) (string, error) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, ".gitlab-ci.yml")
	output := filepath.Join(dir, "out")

	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: path, OutputDirectory: output}); err != nil {
		return "", err
	}

	content, err := os.ReadFile(filepath.Join(output, ".gitlab-ci.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content), nil
}

// TestAliasBombIsRejected covers a denial of service. goccy resolves an
// alias every time it is named and has no cap of its own, so a 380-byte
// file reached 110MB of HCL and 73GB of allocation before this guard.
func TestAliasBombIsRejected(t *testing.T) {
	t.Parallel()

	_, err := unparseGitLab(t, aliasBomb(11, 10))
	if err == nil {
		t.Fatal("expected the alias bomb to be rejected, got no error")
	}

	if !errors.Is(err, errYAMLExhausting) {
		t.Fatalf("expected errYAMLExhausting, got: %v", err)
	}
}

// TestDeepNestingIsRejected covers the other unbounded shape: nesting deep
// enough to exhaust the stack in the recursive walks this package runs.
func TestDeepNestingIsRejected(t *testing.T) {
	t.Parallel()

	deep := "stages: [build]\njob: " +
		strings.Repeat("[", 100000) + strings.Repeat("]", 100000) + "\n"

	_, err := unparseGitLab(t, deep)
	if err == nil {
		t.Fatal("expected the deeply nested document to be rejected, got no error")
	}

	if !errors.Is(err, errYAMLExhausting) {
		t.Fatalf("expected errYAMLExhausting, got: %v", err)
	}
}

// TestModestAliasingStillWorks keeps the guard from being read as a ban on
// anchors. Sharing a block across jobs is the normal way to write GitLab
// YAML and has to keep working.
func TestModestAliasingStillWorks(t *testing.T) {
	t.Parallel()

	hcl, err := unparseGitLab(t, "stages: [build]\n"+
		".base: &base\n  stage: build\n  script: [echo hi]\n"+
		"first:\n  <<: *base\nsecond:\n  <<: *base\n")
	if err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	if count := strings.Count(hcl, "echo hi"); count < 2 {
		t.Fatalf("expected the anchor to resolve for both jobs, got %d occurrences in:\n%s", count, hcl)
	}
}

// TestGuardLeavesOtherErrorsToGoccy keeps the pre-pass from taking over
// reporting. Everything except the two exhaustion shapes has to reach the
// caller with goccy's own message, which the rest of this package and its
// fixtures are written against.
func TestGuardLeavesOtherErrorsToGoccy(t *testing.T) {
	t.Parallel()

	// yaml.v3 rejects a tab as indentation too, so a document it dislikes
	// for some other reason still has to come back as goccy described it.
	_, err := unparseGitLab(t, "stages: [build]\njob:\n\tscript: echo hi\n")
	if err == nil {
		t.Fatal("expected an error for the malformed document, got none")
	}

	if errors.Is(err, errYAMLExhausting) {
		t.Fatalf("a malformed document was reported as exhausting: %v", err)
	}
}
