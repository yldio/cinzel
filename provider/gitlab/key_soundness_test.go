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

// TestInvalidUTF8IsRejected covers a silent rename. goccy reads a byte that
// cannot start a UTF-8 sequence as one anyway and the encoder writes it
// back as U+FFFD, so a job named with such a byte comes back under a
// different name and the pipeline that leaves is not the one that arrived.
func TestInvalidUTF8IsRejected(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		yaml []byte
	}{
		{"job name", []byte("stages: [build]\nb\xff\xfead:\n  stage: build\n  script: [echo hi]\n")},
		{"script value", []byte("stages: [build]\njob:\n  stage: build\n  script: [\"echo \xff\xfe hi\"]\n")},
		{"stage name", []byte("stages: [\"bu\xff\xfeild\"]\njob:\n  stage: \"bu\xff\xfeild\"\n  script: [echo hi]\n")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := unparseGitLab(t, string(tc.yaml))
			if err == nil {
				t.Fatal("expected the invalid UTF-8 to be rejected, got no error")
			}

			if !errors.Is(err, errInvalidUTF8) {
				t.Fatalf("expected errInvalidUTF8, got: %v", err)
			}
		})
	}
}

// TestValidMultibyteStillWorks keeps the check above from being read as a
// ban on anything outside ASCII. A pipeline is free to name things in any
// script, and emoji are four-byte sequences that must survive too.
func TestValidMultibyteStillWorks(t *testing.T) {
	t.Parallel()

	hcl, err := unparseGitLab(t, "stages: [build]\njob:\n  stage: build\n  script: [\"echo héllo 世界 🎉\"]\n")
	if err != nil {
		t.Fatalf("Unparse() error = %v", err)
	}

	if !strings.Contains(hcl, "héllo 世界 🎉") {
		t.Fatalf("expected the multibyte text to survive, HCL was:\n%s", hcl)
	}
}
