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
