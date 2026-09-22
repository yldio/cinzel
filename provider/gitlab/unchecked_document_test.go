// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"strings"
	"testing"
)

// TestDocumentOnlyOneReaderTakesIsRejected covers the case where the pass that
// runs the alias cap, the non-string-key rule and the tag rule cannot read the
// document, but the decoder that writes the HCL can.
//
// Every one of these used to go through with all three checks skipped, because
// yaml.v3's refusal was handed on in the expectation that goccy would report
// the same fault. It does not: on these four it reports nothing at all.
func TestDocumentOnlyOneReaderTakesIsRejected(t *testing.T) {
	t.Parallel()

	const job = "stages: [build]\nbuild:\n  stage: build\n  script: [\"echo a\"]\n"

	for _, tc := range []struct{ name, yaml string }{
		{"unknown directive", "%FOO bar\n---\n" + job},
		{"YAML version this reader does not implement", "%YAML 1.2\n---\n" + job},
		{"undefined tag handle", "stages: [build]\nbuild:\n  stage: build\n  script: !e!foo [a]\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := unparseGitLab(t, tc.yaml)
			if err == nil {
				t.Fatal("expected the unreadable document to be rejected, got no error")
			}

			if !errors.Is(err, errYAMLOnlyOneReaderTakes) {
				t.Fatalf("expected errYAMLOnlyOneReaderTakes, got: %v", err)
			}
		})
	}
}

// TestTheCheckedRulesStillApplyBehindADirective is the reason the case above
// matters: each of these is refused on its own, and each went through when a
// "%FOO" line was put in front of it.
func TestTheCheckedRulesStillApplyBehindADirective(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, yaml string }{
		{
			"a non-string key",
			"stages: [build]\nbuild:\n  stage: build\n  script: [\"echo a\"]\n  variables:\n    ~: x\n",
		},
		{
			"a tag that cannot cross into HCL",
			"stages: [build]\nbuild:\n  stage: build\n  script: !reference [.setup, script]\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := unparseGitLab(t, tc.yaml); err == nil {
				t.Fatal("expected the rule to apply, got no error")
			}

			if _, err := unparseGitLab(t, "%FOO bar\n---\n"+tc.yaml); err == nil {
				t.Fatal("the rule was skipped once a directive was put in front of the document")
			}
		})
	}
}

// TestGoccysOwnRefusalStillReachesTheCaller keeps the guard from taking over
// the reporting of a document goccy refuses too. Its message carries the line
// and column the rest of this package is written against, so it is the one to
// print.
func TestGoccysOwnRefusalStillReachesTheCaller(t *testing.T) {
	t.Parallel()

	_, err := unparseGitLab(t, "stages: [build]\nbuild:\n  stage: build\n  script: [\"a\"]\nbuild:\n  stage: build\n  script: [\"b\"]\n")
	if err == nil {
		t.Fatal("expected the duplicate key to be rejected, got no error")
	}

	if errors.Is(err, errYAMLOnlyOneReaderTakes) {
		t.Fatalf("the guard took over a refusal goccy reports better: %v", err)
	}

	if !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("expected goccy's own message, got: %v", err)
	}
}
