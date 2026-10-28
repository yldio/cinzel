// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A comment on a list-valued attribute used to be emitted against the next key
// down, because yaml.v3 renders a line comment where the node's own line ends
// and a filled block collection ends after its last entry. It has to come back
// as a note on the attribute it was written on, not on its neighbour.
func TestACommentOnAListAttributeSurvivesARoundtrip(t *testing.T) {
	const src = "step \"hi\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n  runs_on {\n    runners = \"ubuntu-latest\"\n  }\n\n" +
		"  steps = [step.hi]\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n\n  name = \"W\"\n\n" +
		"  on \"push\" {\n    tags = [\"v*\"] # only release tags\n  }\n\n" +
		"  jobs = [job.b]\n}\n"

	yaml := parsedYAML(t, src)

	if !strings.Contains(yaml, "tags: # only release tags") {
		t.Errorf("the comment did not land on the key it was written on:\n%s", yaml)
	}

	if !strings.Contains(unparseWorkflow(t, yaml), "# only release tags") {
		t.Error("the comment did not survive the trip back to HCL")
	}
}

// unparseWorkflow converts the YAML back to HCL and returns what it wrote.
func unparseWorkflow(t *testing.T, yaml string) string {
	t.Helper()

	dir := t.TempDir()
	in := filepath.Join(dir, "w.yaml")
	out := filepath.Join(dir, "hcl")

	if err := os.WriteFile(in, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(filepath.Join(out, "w.hcl"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// parsedYAML parses the HCL and returns the workflow it wrote.
func parsedYAML(t *testing.T, hcl string) string {
	t.Helper()

	out := parseWorkflow(t, hcl)

	content, err := os.ReadFile(filepath.Join(out, "w.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	return string(content)
}

// An action's "runs.steps" is a reference list like a job's, so a comment on
// it has the same one place to go. The job path carried it and this one did
// not, which left the same HCL behaving differently either side of a document
// boundary.
func TestAnActionStepsListKeepsItsComment(t *testing.T) {
	const src = "action \"act\" {\n  filename    = \"act\"\n" +
		"  name        = \"Act\"\n  description = \"d\"\n\n" +
		"  runs {\n    using = \"composite\"\n\n" +
		"    # the steps it runs\n    steps = [step.s] # in order\n  }\n}\n\n" +
		"step \"s\" {\n  run   = \"echo hi\"\n  shell = \"bash\"\n}\n"

	out := parseWorkflow(t, src)

	content, err := os.ReadFile(filepath.Join(out, "act", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "# the steps it runs\n  steps: # in order") {
		t.Errorf("the comments did not reach the steps key:\n%s", content)
	}
}

// The point of a comment is that it survives being converted and converted
// back. These are the attributes whose YAML key is not the line the HCL
// attribute was written on — a reference list, a block that collapses to one
// attribute, a matrix axis that becomes a two-attribute block — so each needs
// the writer on the way back to put the comment where the reader will find it.
func TestCommentsSurviveARoundtripWhateverTheyAreWrittenOn(t *testing.T) {
	const src = "step \"s\" {\n  run = \"echo hi\"\n}\n\n" +
		"job \"b\" {\n" +
		"  runs_on {\n    runners = \"ubuntu-latest\" # the runner\n  }\n\n" +
		"  strategy {\n    fail_fast = true # stop early\n\n" +
		"    matrix {\n      variable {\n        name  = \"os\"\n" +
		"        value = [\"ubuntu-latest\"] # axis values\n      }\n    }\n  }\n\n" +
		"  steps = [step.s] # which steps\n}\n\n" +
		"workflow \"w\" {\n  filename = \"w\"\n\n  name = \"W\"\n\n" +
		"  on \"push\" {}\n\n  jobs = [job.b] # which jobs\n}\n"

	yaml := parsedYAML(t, src)
	back := unparseWorkflow(t, yaml)

	for _, want := range []string{
		"# the runner",
		"# stop early",
		"# axis values",
		"# which steps",
		"# which jobs",
	} {
		if !strings.Contains(yaml, want) {
			t.Errorf("want %q in the YAML:\n%s", want, yaml)
		}

		if !strings.Contains(back, want) {
			t.Errorf("want %q back in the HCL:\n%s", want, back)
		}
	}
}

// An action's own attributes travel the same way, and "using" and the branding
// keys are written by a path of their own that took no comment.
func TestActionCommentsSurviveARoundtrip(t *testing.T) {
	const src = "action \"act\" {\n  filename    = \"act\"\n" +
		"  name        = \"Act\" # what it is called\n" +
		"  description = \"d\" # what it does\n\n" +
		"  branding {\n    icon  = \"check\" # the glyph\n" +
		"    color = \"blue\"\n  }\n\n" +
		"  runs {\n    using = \"composite\" # how it runs\n\n" +
		"    steps = [step.s] # in order\n  }\n}\n\n" +
		"step \"s\" {\n  run   = \"echo hi\"\n  shell = \"bash\"\n}\n"

	out := parseWorkflow(t, src)

	content, err := os.ReadFile(filepath.Join(out, "act", "action.yml"))
	if err != nil {
		t.Fatal(err)
	}

	yaml := string(content)
	back := unparseAction(t, yaml)

	for _, want := range []string{
		"# what it is called",
		"# what it does",
		"# the glyph",
		"# how it runs",
		"# in order",
	} {
		if !strings.Contains(yaml, want) {
			t.Errorf("want %q in the YAML:\n%s", want, yaml)
		}

		if !strings.Contains(back, want) {
			t.Errorf("want %q back in the HCL:\n%s", want, back)
		}
	}
}

// unparseAction converts an action YAML back to HCL and returns what it wrote.
func unparseAction(t *testing.T, yaml string) string {
	t.Helper()

	dir := t.TempDir()
	in := filepath.Join(dir, "act", "action.yml")
	out := filepath.Join(dir, "hcl")

	if err := os.MkdirAll(filepath.Dir(in), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(in, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Unparse(provider.ProviderOps{File: in, OutputDirectory: out}); err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".hcl") {
			content, err := os.ReadFile(filepath.Join(out, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}

			return string(content)
		}
	}

	t.Fatal("unparse wrote no HCL")

	return ""
}
