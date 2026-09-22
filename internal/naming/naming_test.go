// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package naming

import (
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
)

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

// TestSanitizeIdentifierOutputIsAnHCLIdentifier walks every rune, because the
// disagreement it guards is a table difference and a handful of samples would
// not find the next one.
//
// unicode.IsLetter accepts runes HCL's own scanner refuses, so a name carrying
// one came back unchanged and was written into a reference such as
// "job.ࡠalpha". HCL reads that as a character it has no use for, and cinzel's
// own parse could not read back the file its unparse had just written with
// exit 0.
func TestSanitizeIdentifierOutputIsAnHCLIdentifier(t *testing.T) {
	for r := rune(1); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue // a surrogate half is not a character
		}

		got := SanitizeIdentifier("a" + string(r))

		if !hclsyntax.ValidIdentifier(got) {
			t.Fatalf("SanitizeIdentifier(%q) returned %q, which HCL refuses as an identifier", "a"+string(r), got)
		}
	}
}

// TestSanitizeIdentifierKeepsASCIIWords pins the common case against the rune
// walk above, which would still pass if every letter were replaced.
func TestSanitizeIdentifierKeepsASCIIWords(t *testing.T) {
	if got := SanitizeIdentifier("Deploy_to_prod2"); got != "Deploy_to_prod2" {
		t.Fatalf("expected the name unchanged, got %q", got)
	}
}
