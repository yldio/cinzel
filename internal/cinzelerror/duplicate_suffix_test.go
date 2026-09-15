// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// An error built through ErrOpenIssue or ProcessHCLDiags already ends with the
// suffix, and New used to append a second copy of it.
func TestOpenIssueSuffixAppearsOnce(t *testing.T) {
	for name, err := range map[string]error{
		"plain":            errors.New("something went wrong"),
		"already suffixed": fmt.Errorf("something went wrong, %w", ErrOpenIssue),
	} {
		t.Run(name, func(t *testing.T) {
			got := New(err).Err.Error()

			if n := strings.Count(got, OpenIssue); n != 1 {
				t.Errorf("want the suffix once, got %d: %s", n, got)
			}
		})
	}
}

// The same holds with a context prefix, which takes the other branch.
func TestOpenIssueSuffixAppearsOnceWithPrefix(t *testing.T) {
	err := fmt.Errorf("something went wrong, %w", ErrOpenIssue)
	got := New(err, "error in file 'x.yaml'").Err.Error()

	if n := strings.Count(got, OpenIssue); n != 1 {
		t.Errorf("want the suffix once, got %d: %s", n, got)
	}

	if !strings.Contains(got, "error in file 'x.yaml'") {
		t.Errorf("want the prefix kept, got: %s", got)
	}
}
