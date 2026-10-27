// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

import (
	"strings"
	"testing"
)

// The project rule is: quote when needed, with double quotes, never relying on
// single quotes. A capitalized boolean went out bare, and a value padded with
// spaces was left to yaml.v3, which reaches for single quotes.
func TestQuotingGaps(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value string
		want  string
	}{
		{name: "capital Yes", value: "Yes", want: `a: "Yes"`},
		{name: "capital Off", value: "Off", want: `a: "Off"`},
		{name: "upper TRUE", value: "TRUE", want: `a: "TRUE"`},
		{name: "single letter y", value: "y", want: `a: "y"`},
		{name: "single letter N", value: "N", want: `a: "N"`},
		{name: "capital Null", value: "Null", want: `a: "Null"`},
		{name: "leading and trailing space", value: " padded ", want: `a: " padded "`},
		{name: "trailing space only", value: "tail ", want: `a: "tail "`},
		{name: "a tab at the end", value: "tab\t", want: "a: \"tab\\t\""},
		{name: "an ordinary word is left alone", value: "yesterday", want: "a: yesterday"},
		{name: "an inner space is fine", value: "two words", want: "a: two words"},

		// An indicator is only an indicator in the first position. Scanning
		// the whole value quoted every expression for the braces in its
		// middle.
		{name: "an expression", value: "${{ github.sha }}", want: "a: ${{ github.sha }}"},
		{name: "braces inside", value: "foo{bar}baz", want: "a: foo{bar}baz"},
		{name: "commas inside", value: "x,y,z", want: "a: x,y,z"},
		{name: "brackets inside", value: "matrix[0]", want: "a: matrix[0]"},
		{name: "a colon with no space", value: "foo:bar", want: "a: foo:bar"},
		{name: "a hash with no space", value: "foo#bar", want: "a: foo#bar"},

		// These still read as something other than the string written.
		{name: "a leading brace", value: "{bar}", want: `a: "{bar}"`},
		{name: "a leading bracket", value: "[bar]", want: `a: "[bar]"`},
		{name: "a leading ampersand", value: "&anchor", want: `a: "&anchor"`},
		{name: "a leading asterisk", value: "*alias", want: `a: "*alias"`},
		{name: "a leading bang", value: "!tag", want: `a: "!tag"`},
		{name: "a leading pipe", value: "|block", want: `a: "|block"`},
		{name: "a leading hash", value: "#comment", want: `a: "#comment"`},
		{name: "a colon then a space", value: "foo: bar", want: `a: "foo: bar"`},
		{name: "a trailing colon", value: "trailing:", want: `a: "trailing:"`},
		{name: "a space then a hash", value: "foo # bar", want: `a: "foo # bar"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := New()
			doc.Set("a", Scalar(tc.value))

			out, err := Encode(doc)
			if err != nil {
				t.Fatal(err)
			}

			got := strings.TrimSpace(string(out))

			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}

			if strings.Contains(got, "'") {
				t.Errorf("single quotes in %q", got)
			}
		})
	}
}
