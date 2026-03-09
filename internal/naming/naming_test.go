// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package naming

import "testing"

func TestSanitizeIdentifier(t *testing.T) {

	if got := SanitizeIdentifier("build-test"); got != "build_test" {
		t.Fatalf("expected build_test, got %s", got)
	}

	if got := SanitizeIdentifier("123abc"); got != "_123abc" {
		t.Fatalf("expected _123abc, got %s", got)
	}
}

func TestUniqueIdentifier(t *testing.T) {

	if got := UniqueIdentifier("job", []string{"build", "test"}); got != "job" {
		t.Fatalf("expected job, got %s", got)
	}

	if got := UniqueIdentifier("job", []string{"job", "job_2"}); got != "job_3" {
		t.Fatalf("expected job_3, got %s", got)
	}
}

func TestKeyMapping(t *testing.T) {

	if got := ToHCLKey("runs-on"); got != "runs_on" {
		t.Fatalf("expected runs_on, got %s", got)
	}

	if got := ToYAMLKey("runs_on"); got != "runs-on" {
		t.Fatalf("expected runs-on, got %s", got)
	}
}
