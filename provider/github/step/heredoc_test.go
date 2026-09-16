// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// hclwrite.Format used to rewrite a run script as if it were HCL, because the
// heredoc was built from quote tokens and its "${" opened template mode.
func TestSetAsHeredocSurvivesFormatting(t *testing.T) {
	script := strings.Join([]string{
		`tag="${{ github.event.inputs.tag }}"`,
		`tag="${tag#v}"`,
		`if [ -n "$tag" ] && [ "$bump" != "none" ]; then`,
		`  echo "tag=$tag" >> "$GITHUB_OUTPUT"`,
		`fi`,
	}, "\n")

	file := hclwrite.NewEmptyFile()
	file.Body().AppendNewBlock("step", []string{"s"}).Body().
		SetAttributeRaw("run", setAsHeredoc(script))

	got := string(hclwrite.Format(file.Bytes()))

	// Spacing the formatter used to introduce, each of which breaks the shell.
	for _, broken := range []string{"${ {", "> >", "if[", "] ;", "tag #v"} {
		if strings.Contains(got, broken) {
			t.Errorf("formatter introduced %q:\n%s", broken, got)
		}
	}

	for _, want := range []string{
		`if [ -n "$tag" ] && [ "$bump" != "none" ]; then`,
		`  echo "tag=$tag" >> "$GITHUB_OUTPUT"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q:\n%s", want, got)
		}
	}

	// Template markers are doubled, which is how parse reads them back.
	if !strings.Contains(got, `$${{ github.event.inputs.tag }}`) {
		t.Errorf("expression was not escaped:\n%s", got)
	}
}

func TestEscapeTemplateMarkers(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`a="${{ x }}"`, `a="$${{ x }}"`},
		{`a="${b#v}"`, `a="$${b#v}"`},
		{`%{ if x }`, `%%{ if x }`},
		{`plain $VAR and 50% done`, `plain $VAR and 50% done`},
	} {
		if got := escapeTemplateMarkers(tc.in); got != tc.want {
			t.Errorf("escapeTemplateMarkers(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFreeHeredocMarker(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []string
		want  string
	}{
		{"no clash", []string{"echo hi"}, "EOF"},
		{"body writes its own heredoc", []string{"cat <<EOF", "x", "EOF"}, "EOF_1"},
		{"indented marker still clashes", []string{"  EOF"}, "EOF_1"},
		{"walks past a taken suffix", []string{"EOF", "EOF_1", "EOF_2"}, "EOF_3"},
		{"a marker inside a line is not a clash", []string{"echo EOF now"}, "EOF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := freeHeredocMarker(tc.lines); got != tc.want {
				t.Errorf("freeHeredocMarker(%q) = %q, want %q", tc.lines, got, tc.want)
			}
		})
	}
}

// A script writing its own "EOF" heredoc used to close ours early, leaving HCL
// where the rest of the script was read as block syntax.
func TestSetAsHeredocPicksAFreeMarker(t *testing.T) {
	script := strings.Join([]string{"cat > cfg <<EOF", "key: value", "EOF", "echo done"}, "\n")

	file := hclwrite.NewEmptyFile()
	file.Body().AppendNewBlock("step", []string{"s"}).Body().
		SetAttributeRaw("run", setAsHeredoc(script))

	got := string(hclwrite.Format(file.Bytes()))

	if !strings.Contains(got, "<<-EOF_1\n") || !strings.Contains(got, "\nEOF_1\n") {
		t.Errorf("heredoc did not move off the clashing marker:\n%s", got)
	}

	// The body's own marker must survive untouched.
	if !strings.Contains(got, "cat > cfg <<EOF\n") {
		t.Errorf("body marker was rewritten:\n%s", got)
	}
}
