// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package unescape

import "testing"

func TestUnicode(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"emoji", `name: "\U0001F46E Lint"`, "name: \"\U0001F46E Lint\""},
		{"zwj", `name = "a‍b"`, "name = \"a‍b\""},
		{"escaped backslash is text", `run: "echo caf\\u00e9"`, `run: "echo caf\\u00e9"`},
		{"literal backslash then escape", `run: "caf\\é"`, "run: \"caf\\\\é\""},
		{"control stays escaped", `v: "	"`, `v: "	"`},
		{"not an escape", `v: "u0041"`, `v: "u0041"`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := string(Unicode([]byte(tc.in))); got != tc.want {
				t.Errorf("Unicode(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
