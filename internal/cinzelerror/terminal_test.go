// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package cinzelerror

import (
	"strings"
	"testing"
)

// A job name is quoted back in several validation errors, so a name carrying an
// ANSI escape sequence used to reach the terminal intact and be acted on there.
func TestSafeForTerminalEscapesControlCharacters(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{
			name: "ansi colour sequence",
			in:   "jobs.\x1b[31mPWNED\x1b[0m",
			want: "jobs.\\u001b[31mPWNED\\u001b[0m",
		},
		{
			name: "carriage return returns to the start of the line",
			in:   "job\rforged",
			want: "job\\u000dforged",
		},
		{
			name: "bell and backspace",
			in:   "job\a\b",
			want: "job\\u0007\\u0008",
		},
		{
			name: "c1 control sequence introducer",
			in:   "job" + string(rune(0x9b)) + "31m",
			want: "job\\u009b31m",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := SafeForTerminal(tc.in); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// Rewriting more than the control characters would mangle every error that
// quotes a name, so the common case has to come back untouched.
func TestSafeForTerminalKeepsPrintableText(t *testing.T) {
	for _, in := range []string{
		"",
		"jobs.build-and-test: 'runs-on' is required",
		"jobs.本番デプロイ.permissions",
		"jobs.déployer: invalid",
		"jobs.build\n  1 | jobs:\n\tindented",
	} {
		if got := SafeForTerminal(in); got != in {
			t.Errorf("SafeForTerminal(%q) = %q, want it unchanged", in, got)
		}
	}
}

// The YAML decoder quotes the offending source across several lines, so a
// message keeps its newlines while the escape inside it is rewritten.
func TestSafeForTerminalKeepsLayoutOfQuotedSource(t *testing.T) {
	in := "workflow_yaml: [2:15] string was used where mapping is expected\n" +
		"   1 | jobs:\n>  2 |   \x1b[31mX\x1b[0m: scalar\n"

	got := SafeForTerminal(in)

	if strings.ContainsRune(got, '\x1b') {
		t.Errorf("escape survived: %q", got)
	}

	if strings.Count(got, "\n") != strings.Count(in, "\n") {
		t.Errorf("newlines changed: got %d, want %d", strings.Count(got, "\n"), strings.Count(in, "\n"))
	}
}
