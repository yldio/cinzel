// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"strings"
	"testing"
)

// TestNonStringMappingKeyIsRejected covers silent corruption. goccy renders
// a key by its text, so "~", "null" and "NULL" all arrive as "null" and
// overwrite one another, and the HCL that comes out reads back as the
// string "null", which is a different pipeline.
func TestNonStringMappingKeyIsRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, yaml string }{
		{
			"null key",
			"stages: [build]\njob:\n  stage: build\n  script: [echo hi]\n  variables:\n    ~: x\n",
		},
		{
			"three spellings of null collapse to one",
			"stages: [build]\njob:\n  stage: build\n  script: [echo hi]\n  variables:\n    ~: a\n    null: b\n    NULL: c\n",
		},
		{
			"bool key",
			"stages: [build]\njob:\n  stage: build\n  script: [echo hi]\n  variables:\n    true: x\n",
		},
		{
			"int key",
			"stages: [build]\njob:\n  stage: build\n  script: [echo hi]\n  variables:\n    123: x\n",
		},
		{
			"float key",
			"stages: [build]\njob:\n  stage: build\n  script: [echo hi]\n  variables:\n    1.50: x\n",
		},
		{
			"null job name",
			"stages: [build]\n~:\n  stage: build\n  script: [echo hi]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := unparseGitLab(t, tc.yaml)
			if err == nil {
				t.Fatal("expected the non-string key to be rejected, got no error")
			}

			if !errors.Is(err, errNonStringKey) {
				t.Fatalf("expected errNonStringKey, got: %v", err)
			}
		})
	}
}

// TestQuotedKeysThatLookLikeScalarsStillWork keeps the check above from
// being read as a ban on those names. Quoted, they are strings, and a
// pipeline is free to use them. HCL writes null and true unquoted, so what
// matters is that the values stay attached to the right keys and the names
// come back quoted.
func TestQuotedKeysThatLookLikeScalarsStillWork(t *testing.T) {
	t.Parallel()

	hcl, err := unparseGitLab(t, "stages: [build]\njob:\n  stage: build\n  script: [echo hi]\n"+
		"  variables:\n    \"null\": a\n    \"true\": b\n    \"123\": c\n")
	if err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	for _, want := range []string{`null  = "a"`, `true  = "b"`, `"123" = "c"`} {
		if !strings.Contains(hcl, want) {
			t.Errorf("expected %s in the HCL, got:\n%s", want, hcl)
		}
	}
}

// TestMergeKeysStillWork keeps the check from rejecting the merge key,
// which is how a pipeline shares a block between jobs.
func TestMergeKeysStillWork(t *testing.T) {
	t.Parallel()

	hcl, err := unparseGitLab(t, "stages: [build]\n"+
		".base: &base\n  stage: build\n  script: [echo hi]\n"+
		"first:\n  <<: *base\n")
	if err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	if !strings.Contains(hcl, "echo hi") {
		t.Fatalf("expected the merged block to reach the job, HCL was:\n%s", hcl)
	}
}

// TestEmptyJobNameIsRejected covers a job that cannot be referred to by
// needs or extends and has no key to emit. Unparse used to accept it and
// write id = "", which the parse direction then refused, so the pipeline
// could not come back.
func TestEmptyJobNameIsRejected(t *testing.T) {
	t.Parallel()

	_, err := unparseGitLab(t, "stages: [build]\n\"\":\n  stage: build\n  script: [echo hi]\n")
	if err == nil {
		t.Fatal("expected the unnamed job to be rejected, got no error")
	}

	if !errors.Is(err, errBlockIDNotString) {
		t.Fatalf("expected errBlockIDNotString, got: %v", err)
	}
}
