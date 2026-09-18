// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

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

func TestUniqueIdentifierInSet(t *testing.T) {
	free := map[string]struct{}{"build": {}, "test": {}}

	if got := UniqueIdentifierInSet("job", free); got != "job" {
		t.Fatalf("expected job, got %s", got)
	}

	taken := map[string]struct{}{"job": {}, "job_2": {}}

	if got := UniqueIdentifierInSet("job", taken); got != "job_3" {
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

// A digit outside ASCII is still a digit HCL refuses to start an identifier
// with. The check read out[0], a byte, which on a multibyte rune is a UTF-8
// lead byte and never a digit, so the prefix was skipped and the name went out
// unparseable.
func TestSanitizeIdentifierPrefixesEveryLeadingDigit(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "an ascii digit", in: "3d", want: "_3d"},
		{name: "an arabic-indic digit", in: "٣build", want: "_٣build"},
		{name: "a fullwidth digit", in: "３d", want: "_３d"},
		{name: "a letter outside ascii is left alone", in: "café", want: "café"},
		{name: "a digit anywhere else is left alone", in: "go2", want: "go2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SanitizeIdentifier(tc.in); got != tc.want {
				t.Errorf("SanitizeIdentifier(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
