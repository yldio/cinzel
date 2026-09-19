// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"strings"
	"testing"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/cinzel/internal/hclparser"
)

// parseSteps runs the step blocks in src through the same path the provider
// does, and returns whatever the parse refused.
func parseSteps(t *testing.T, src string) error {
	t.Helper()

	var cfg struct {
		Steps StepListConfig `hcl:"step,block"`
	}

	parser := hclparse.NewParser()

	hclFile, diags := parser.ParseHCL([]byte(src), "example.hcl")

	if diags.HasErrors() {
		t.Fatalf("parsing the fixture: %v", diags)
	}

	if diags := gohcl.DecodeBody(hclFile.Body, nil, &cfg); diags.HasErrors() {
		t.Fatalf("decoding the fixture: %v", diags)
	}

	_, err := cfg.Steps.Parse(hclparser.NewHCLVars())

	return err
}

// A type check reports the type it wanted and the one it found and nothing
// more, so a run over a directory said only "unsupported type, expected
// string, found number" for any of a dozen steps and four attributes.
func TestAStepAttributeErrorNamesTheStepAndTheAttribute(t *testing.T) {
	for _, tc := range []struct {
		name string
		attr string
		body string
	}{
		{name: "id", attr: "id", body: `id = 42`},
		{name: "ignore_id", attr: "ignore_id", body: `ignore_id = "yes"`},
		{name: "if", attr: "if", body: `if = 42`},
		{name: "name", attr: "name", body: `name = 42`},
		{name: "run", attr: "run", body: `run = 42`},
		{name: "working_directory", attr: "working_directory", body: `working_directory = 42`},
		{name: "shell", attr: "shell", body: `shell = 42`},
		{name: "timeout_minutes", attr: "timeout_minutes", body: `timeout_minutes = "ten"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := parseSteps(t, "step \"mystep\" {\n  "+tc.body+"\n}\n")

			if err == nil {
				t.Fatal("expected an error")
			}

			got := err.Error()

			if !strings.Contains(got, "'mystep'") {
				t.Errorf("error does not name the step: %v", err)
			}

			if !strings.Contains(got, tc.attr) {
				t.Errorf("error does not name the attribute %q: %v", tc.attr, err)
			}
		})
	}
}

func TestAStepWithNothingWrongWithItStillParses(t *testing.T) {
	src := "step \"mystep\" {\n" +
		"  id                = \"mystep\"\n" +
		"  if                = \"$${{ success() }}\"\n" +
		"  name              = \"a name\"\n" +
		"  run               = \"echo a\"\n" +
		"  working_directory = \"./sub\"\n" +
		"  shell             = \"bash\"\n" +
		"  timeout_minutes   = 5\n" +
		"  continue_on_error = true\n" +
		"}\n"

	if err := parseSteps(t, src); err != nil {
		t.Fatalf("expected the step to parse, got %v", err)
	}
}
