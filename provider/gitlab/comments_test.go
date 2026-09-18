// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// Comments were read nowhere in this provider, so every one written in an HCL
// pipeline was dropped on the way to YAML.
func TestCommentsSurviveParse(t *testing.T) {
	for _, tc := range []struct {
		name string
		hcl  string
		want string
	}{
		{
			name: "above a top-level attribute",
			hcl:  "# the stages we run\nstages = [\"build\"]\n" + commentJobHCL(""),
			want: "# the stages we run\nstages:",
		},
		{
			name: "trailing a top-level attribute",
			hcl:  "stages = [\"build\"] # in order\n" + commentJobHCL(""),
			want: "stages: # in order",
		},
		{
			name: "above a job block",
			hcl:  "stages = [\"build\"]\n# the build job\n" + commentJobHCL(""),
			want: "# the build job\nbuild:",
		},
		{
			name: "above a job attribute",
			hcl:  "stages = [\"build\"]\n" + commentJobHCL("  # what it runs\n  script = [\"make\"]\n"),
			want: "  # what it runs\n  script:",
		},
		{
			name: "trailing a job attribute",
			hcl:  "stages = [\"build\"]\n" + commentJobHCL("  script = [\"make\"] # one command\n"),
			want: "script: # one command",
		},
		{
			name: "closing a job block",
			hcl:  "stages = [\"build\"]\n" + commentJobHCL("  script = [\"make\"]\n  # that is the whole job\n"),
			want: "  # that is the whole job",
		},
		{
			// A block written once is a mapping under its own key, so the
			// comment goes above that key.
			name: "above a block written once",
			hcl: "stages = [\"build\"]\n" + commentJobHCL(
				"  script = [\"make\"]\n\n  # keep the binary\n  artifacts {\n    paths = [\"bin/\"]\n  }\n"),
			want: "  # keep the binary\n  artifacts:",
		},
		{
			name: "closing a block written once",
			hcl: "stages = [\"build\"]\n" + commentJobHCL(
				"  script = [\"make\"]\n\n  artifacts {\n    paths = [\"bin/\"]\n    # nothing else\n  }\n"),
			want: "    # nothing else",
		},
		{
			// A block written more than once is a list, so the comment goes
			// above the entry it was written above rather than above the key.
			name: "above one entry of a list",
			hcl: "stages = [\"build\"]\n" + commentJobHCL(
				"  script = [\"make\"]\n\n  # only on main\n  rule {\n    if = \"a\"\n  }\n\n  rule {\n    when = \"never\"\n  }\n"),
			want: "    # only on main\n    - if: a",
		},
		{
			// A variable carrying nothing beyond its value collapses to a
			// plain scalar, so its block's comments have no keys left to sit
			// on and belong to the key it collapsed to.
			name: "inside a collapsed variable block",
			hcl: "stages = [\"build\"]\nvariable \"v\" {\n  # what it holds\n  name  = \"REGISTRY\"\n" +
				"  value = \"example.com\" # the registry\n}\n" + commentJobHCL(""),
			want: "  # what it holds\n  REGISTRY: example.com # the registry",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := parsePipeline(t, tc.hcl); !strings.Contains(got, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, got)
			}
		})
	}
}

// A comment reaching YAML in the wrong place is worse than one dropped: it
// says something about a key its author never wrote it against.
func TestCommentsLandOnlyWhereTheyWereWritten(t *testing.T) {
	const src = "stages = [\"build\"]\n" +
		"job \"build\" {\n  stage  = \"build\"\n  script = [\"make\"] # one command\n}\n"

	got := parsePipeline(t, src)

	if !strings.Contains(got, "script: # one command") {
		t.Errorf("the comment did not land on its own key:\n%s", got)
	}

	if strings.Contains(got, "stage: build # one command") {
		t.Errorf("the comment landed on a key it was not written against:\n%s", got)
	}
}

// commentJobHCL returns a job block carrying the given body lines, which is a
// valid pipeline on its own.
func commentJobHCL(body string) string {
	if body == "" {
		body = "  script = [\"make\"]\n"
	}

	return "job \"build\" {\n  stage = \"build\"\n" + body + "}\n"
}

// parsePipeline converts an HCL pipeline and returns the YAML it produced.
func parsePipeline(t *testing.T, src string) string {
	t.Helper()

	dir := t.TempDir()
	in := filepath.Join(dir, "p.hcl")
	out := filepath.Join(dir, "yaml")

	if err := os.WriteFile(in, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatalf("Parse() error = %v\nHCL:\n%s", err, src)
	}

	content, err := os.ReadFile(filepath.Join(out, ".gitlab-ci.yml"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// The other direction of the roundtrip in unparse_comments_test.go: a comment
// written in HCL has to come back to the HCL it started in, not only reach the
// YAML once.
func TestACommentSurvivesAGitLabRoundtripFromHCL(t *testing.T) {
	const src = "stages = [\"build\"]\n\n" +
		"# the build job\n" +
		"job \"build\" {\n" +
		"  stage = \"build\"\n" +
		"  # what it runs\n" +
		"  script = [\"make\"] # one command\n" +
		"  # that is the whole job\n" +
		"}\n"

	got := unparsePipeline(t, parsePipeline(t, src))

	for _, want := range []string{
		"# the build job",
		"# what it runs",
		"# one command",
		"# that is the whole job",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("want %q after a roundtrip, got:\n%s", want, got)
		}
	}
}

// A comment is prose its author spaced as they chose, so both hops carry the
// bytes rather than a tidied version of them.
func TestAGitLabCommentIsVerbatim(t *testing.T) {
	const marker = "##no space and   wide   spacing"

	fromHCL := parsePipeline(t, commentJobHCL("  "+marker+"\n  script = [\"make\"]\n"))

	if !strings.Contains(fromHCL, marker) {
		t.Errorf("the comment was rewritten on the way to YAML:\n%s", fromHCL)
	}

	fromYAML := unparsePipeline(t, "stages:\n  - build\n\nbuild:\n  stage: build\n  "+marker+"\n  script:\n    - make\n")

	if !strings.Contains(fromYAML, marker) {
		t.Errorf("the comment was rewritten on the way to HCL:\n%s", fromYAML)
	}
}
