// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// Every CLI failure used to end with the invitation to open an issue, so a
// misspelled attribute or a duplicate label told the author to file a bug
// against cinzel over something only they could fix.
func TestUserInputErrorsDropTheIssueLine(t *testing.T) {
	marked := UserInput(errors.New("two step blocks share a label"))

	for _, tc := range []struct {
		name string
		got  string
	}{
		{"on its own", New(marked).Err.Error()},
		{"with a context prefix", New(marked, "error in file 'ci.hcl'").Err.Error()},
		{"wrapped in more context", New(fmt.Errorf("in job 'build': %w", marked)).Err.Error()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Contains(tc.got, OpenIssue) {
				t.Errorf("want no issue line, got %q", tc.got)
			}

			if !strings.Contains(tc.got, "two step blocks share a label") {
				t.Errorf("want the cause kept, got %q", tc.got)
			}
		})
	}
}

// Anything not marked is cinzel's to answer for, and still says so.
func TestUnmarkedErrorsKeepTheIssueLine(t *testing.T) {
	got := New(errors.New("unsupported body type")).Err.Error()

	if !strings.Contains(got, OpenIssue) {
		t.Errorf("want the issue line, got %q", got)
	}
}

// The mark sits on the sentinel at its declaration, so it has to survive the
// fmt.Errorf("%w") wrapping every call site adds, and leave errors.Is against
// that same sentinel working.
func TestTheMarkSurvivesWrapping(t *testing.T) {
	sentinel := UserInput(errors.New("a mapping key must be a string"))
	wrapped := fmt.Errorf("error in file 'ci.yaml': %w", fmt.Errorf("in job 'build': %w", sentinel))

	if !IsUserInput(wrapped) {
		t.Error("want the mark to survive two layers of wrapping")
	}

	if !errors.Is(wrapped, sentinel) {
		t.Error("want errors.Is against the sentinel to still match")
	}
}

// A prefix with no error behind it is the internal case and keeps the line.
func TestNoErrorStillReportsTheIssueLine(t *testing.T) {
	if got := New(nil, "something").Err.Error(); !strings.Contains(got, OpenIssue) {
		t.Errorf("want the issue line, got %q", got)
	}
}
