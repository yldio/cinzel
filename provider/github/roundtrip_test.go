// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/yldio/cinzel/provider"
)

func TestWorkflowRoundtripFixtures(t *testing.T) {
	fixtures := []string{
		"basic_workflow",
		"workflow_call",
		"reusable_job",
		"schedule_workflow",
		"permissions_concurrency",
		"container_services",
		"runs_on_list",
		"local_action",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			tmpDir := t.TempDir()

			inputHCL := filepath.Join("testdata", "fixtures", "workflows", fixture+".hcl")
			parse1Dir := filepath.Join(tmpDir, "parse1")
			unparseDir := filepath.Join(tmpDir, "unparse")
			parse2Dir := filepath.Join(tmpDir, "parse2")

			p := New()

			if err := p.Parse(provider.ProviderOps{File: inputHCL, OutputDirectory: parse1Dir}); err != nil {
				t.Fatal(err)
			}

			yamlFile := filepath.Join(parse1Dir, fixture+".yaml")

			if err := p.Unparse(provider.ProviderOps{File: yamlFile, OutputDirectory: unparseDir}); err != nil {
				t.Fatal(err)
			}

			hclFile := filepath.Join(unparseDir, fixture+".hcl")

			if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parse2Dir}); err != nil {
				t.Fatal(err)
			}

			first, err := os.ReadFile(yamlFile)
			if err != nil {
				t.Fatal(err)
			}

			second, err := os.ReadFile(filepath.Join(parse2Dir, fixture+".yaml"))
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLValueEqual(t, first, second)
		})
	}
}

func assertYAMLValueEqual(t *testing.T, expected []byte, got []byte) {
	t.Helper()

	var expectedValue any

	if err := yaml.Unmarshal(expected, &expectedValue); err != nil {
		t.Fatalf("failed to unmarshal expected YAML: %v", err)
	}

	var gotValue any

	if err := yaml.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("failed to unmarshal generated YAML: %v", err)
	}

	if !reflect.DeepEqual(expectedValue, gotValue) {
		t.Fatalf("roundtrip YAML mismatch\n--- expected ---\n%s\n--- got ---\n%s", string(expected), string(got))
	}
}

func TestActionRoundtripFixtures(t *testing.T) {
	fixtures := []string{
		"composite_action",
		"node_action",
		"docker_action",
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			tmpDir := t.TempDir()

			inputHCL := filepath.Join("testdata", "fixtures", "actions", fixture+".hcl")
			parse1Dir := filepath.Join(tmpDir, "parse1")
			unparseDir := filepath.Join(tmpDir, "unparse")
			parse2Dir := filepath.Join(tmpDir, "parse2")

			p := New()

			if err := p.Parse(provider.ProviderOps{File: inputHCL, OutputDirectory: parse1Dir}); err != nil {
				t.Fatal(err)
			}

			actionYAML := filepath.Join(parse1Dir, fixture, "action.yml")

			if err := p.Unparse(provider.ProviderOps{File: actionYAML, OutputDirectory: unparseDir}); err != nil {
				t.Fatal(err)
			}

			hclFile := filepath.Join(unparseDir, "action.hcl")

			if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parse2Dir}); err != nil {
				t.Fatal(err)
			}

			first, err := os.ReadFile(actionYAML)
			if err != nil {
				t.Fatal(err)
			}

			second, err := os.ReadFile(filepath.Join(parse2Dir, "action", "action.yml"))
			if err != nil {
				t.Fatal(err)
			}

			assertYAMLValueEqual(t, first, second)
		})
	}
}

func TestWorkflowExpressionRoundtripStability(t *testing.T) {
	tmpDir := t.TempDir()

	inputHCL := filepath.Join(tmpDir, "expressions.hcl")
	parse1Dir := filepath.Join(tmpDir, "parse1")
	unparseDir := filepath.Join(tmpDir, "unparse")
	parse2Dir := filepath.Join(tmpDir, "parse2")

	content := `step "build" {
  if = "$${{ failure() }}"
  run = "echo $${{ github.ref }}"
  with {
    name = "token"
    value = "$${{ secrets.GITHUB_TOKEN }}"
  }
  env {
    name = "REF"
    value = "$${{ github.ref }}"
  }
  continue_on_error = "$${{ matrix.experimental }}"
}

job "release" {
  if = "$${{ github.ref == 'refs/heads/main' }}"
  runs_on {
    runners = "$${{ matrix.os }}"
  }

  concurrency {
    group = "$${{ github.workflow }}-$${{ github.ref }}"
    cancel_in_progress = true
  }

  steps = [step.build]
}

workflow "wf" {
  filename = "expressions"
  run_name = "$${{ github.workflow }} #$${{ github.run_number }}"
  on "push" {}
  jobs = [job.release]
}
`

	if err := os.WriteFile(inputHCL, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Parse(provider.ProviderOps{File: inputHCL, OutputDirectory: parse1Dir}); err != nil {
		t.Fatal(err)
	}

	yamlFile := filepath.Join(parse1Dir, "expressions.yaml")

	if err := p.Unparse(provider.ProviderOps{File: yamlFile, OutputDirectory: unparseDir}); err != nil {
		t.Fatal(err)
	}

	hclFile := filepath.Join(unparseDir, "expressions.hcl")
	hclBytes, err := os.ReadFile(hclFile)
	if err != nil {
		t.Fatal(err)
	}

	hclOut := string(hclBytes)

	if !strings.Contains(hclOut, `run_name = "$${{ github.workflow }} #$${{ github.run_number }}"`) {
		t.Fatalf("expected escaped workflow expression in HCL, got:\n%s", hclOut)
	}

	if !strings.Contains(hclOut, `continue_on_error = "$${{ matrix.experimental }}"`) {
		t.Fatalf("expected escaped step expression in HCL, got:\n%s", hclOut)
	}

	if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parse2Dir}); err != nil {
		t.Fatal(err)
	}

	first, err := os.ReadFile(yamlFile)
	if err != nil {
		t.Fatal(err)
	}

	second, err := os.ReadFile(filepath.Join(parse2Dir, "expressions.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	assertYAMLValueEqual(t, first, second)
}

// TestEmojiRoundtripStability verifies that workflow names containing emoji
// ZWJ sequences (e.g. 👮‍♂️) round-trip through YAML → HCL → YAML without
// acquiring Unicode escape sequences in either the HCL or the output YAML.
func TestEmojiRoundtripStability(t *testing.T) {
	tmpDir := t.TempDir()

	inputYAML := filepath.Join(tmpDir, "lint.yaml")
	unparseDir := filepath.Join(tmpDir, "unparse")
	parse2Dir := filepath.Join(tmpDir, "parse2")

	// The workflow name contains a ZWJ emoji sequence: 👮 (U+1F46E) + ZWJ
	// (U+200D) + ♂ (U+2642) + VS-16 (U+FE0F).  hclwrite escapes U+200D as
	// \u200d and yaml.v3 escapes supplementary-plane chars as \UXXXXXXXX, so
	// without the fix both representations acquire Unicode escapes.
	emojiName := "\U0001F46E\u200D\u2642\uFE0F Lint"
	content := "name: " + emojiName + "\non:\n  push:\njobs:\n  lint:\n    runs-on: ubuntu-latest\n    steps:\n      - id: lint\n        name: Run lint\n        run: echo lint\n"

	if err := os.WriteFile(inputYAML, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p := New()

	if err := p.Unparse(provider.ProviderOps{File: inputYAML, OutputDirectory: unparseDir}); err != nil {
		t.Fatal(err)
	}

	hclFile := filepath.Join(unparseDir, "lint.hcl")
	hclBytes, err := os.ReadFile(hclFile)

	if err != nil {
		t.Fatal(err)
	}

	// HCL must not contain Unicode escape sequences for the emoji ZWJ.
	hclStr := string(hclBytes)

	if strings.Contains(hclStr, `\u200d`) || strings.Contains(hclStr, `\u200D`) {
		t.Errorf("HCL output contains escaped ZWJ (\\u200d); want raw UTF-8:\n%s", hclStr)
	}

	if err := p.Parse(provider.ProviderOps{File: hclFile, OutputDirectory: parse2Dir}); err != nil {
		t.Fatal(err)
	}

	roundtripped, err := os.ReadFile(filepath.Join(parse2Dir, "lint.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Round-tripped YAML must not contain Unicode escape sequences.
	roundStr := string(roundtripped)

	if strings.Contains(roundStr, `\U`) || strings.Contains(roundStr, `\u`) {
		t.Errorf("round-tripped YAML contains Unicode escapes; want raw UTF-8:\n%s", roundStr)
	}

	// The emoji name must be preserved verbatim in the output.
	if !strings.Contains(roundStr, emojiName) {
		t.Errorf("round-tripped YAML does not contain emoji name %q:\n%s", emojiName, roundStr)
	}
}
