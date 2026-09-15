// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yldio/cinzel/provider"
)

// A filename is joined onto the output directory and written, so one carrying
// "../" walked out of the directory the caller asked for, and an absolute one
// ignored it entirely. Either way the file landed somewhere nobody asked for
// and nothing said so.
func TestFilenameCannotEscapeTheOutputDirectory(t *testing.T) {
	for _, tc := range []struct {
		name     string
		filename string
	}{
		{name: "parent", filename: "../escaped"},
		{name: "deep parent", filename: "../../../../../../escaped"},
		{name: "parent after a subdirectory", filename: "sub/../../escaped"},
		{name: "absolute", filename: "/tmp/escaped"},
		{name: "bare parent", filename: ".."},
	} {
		t.Run("workflow "+tc.name, func(t *testing.T) {
			assertEscapeRefused(t, workflowHCL(tc.filename))
		})

		t.Run("action "+tc.name, func(t *testing.T) {
			assertEscapeRefused(t, actionHCL(tc.filename))
		})
	}
}

// A filename naming a subdirectory stays inside the output directory, and an
// action already depends on one, so the guard has to let it through.
func TestFilenameMayNameASubdirectory(t *testing.T) {
	for _, tc := range []struct {
		name     string
		filename string
		want     string
	}{
		{name: "plain", filename: "build", want: "build.yaml"},
		{name: "subdirectory", filename: "nested/build", want: filepath.Join("nested", "build.yaml")},
		{name: "dot prefix", filename: "./build", want: "build.yaml"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := parseWorkflow(t, workflowHCL(tc.filename))

			if _, err := os.Stat(filepath.Join(out, tc.want)); err != nil {
				t.Errorf("want %s written, got %v", tc.want, err)
			}
		})
	}
}

// The guard has to refuse before anything is written, not after: a file left
// behind outside the output directory is the whole defect.
func TestRefusedFilenameWritesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "w.hcl")
	output := filepath.Join(dir, "out")

	if err := os.WriteFile(path, []byte(workflowHCL("../escaped")), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: output}); err == nil {
		t.Fatal("want an error, got none")
	}

	if _, err := os.Stat(filepath.Join(dir, "escaped.yaml")); !os.IsNotExist(err) {
		t.Errorf("a file was written outside the output directory: %v", err)
	}
}

// assertEscapeRefused parses the HCL and requires errFilenameEscapes back.
func assertEscapeRefused(t *testing.T, hcl string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "in.hcl")

	if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: filepath.Join(dir, "out")})

	if !errors.Is(err, errFilenameEscapes) {
		t.Errorf("want errFilenameEscapes, got %v", err)
	}
}

// parseWorkflow parses the HCL and returns the output directory it wrote to.
func parseWorkflow(t *testing.T, hcl string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "in.hcl")
	output := filepath.Join(dir, "out")

	if err := os.WriteFile(path, []byte(hcl), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := New().Parse(provider.ProviderOps{File: path, OutputDirectory: output}); err != nil {
		t.Fatal(err)
	}

	return output
}

// workflowHCL returns a minimal workflow carrying filename.
func workflowHCL(filename string) string {
	return "step \"hi\" {\n" +
		"  run = \"echo hi\"\n" +
		"}\n\n" +
		"job \"b\" {\n" +
		"  runs_on {\n" +
		"    runners = \"ubuntu-latest\"\n" +
		"  }\n" +
		"  steps = [step.hi]\n" +
		"}\n\n" +
		"workflow \"w\" {\n" +
		"  filename = \"" + filename + "\"\n" +
		"  name     = \"W\"\n" +
		"  on \"push\" {}\n" +
		"  jobs = [job.b]\n" +
		"}\n"
}

// actionHCL returns a minimal action carrying filename.
func actionHCL(filename string) string {
	return "step \"hi\" {\n" +
		"  run = \"echo hi\"\n" +
		"}\n\n" +
		"action \"a\" {\n" +
		"  filename    = \"" + filename + "\"\n" +
		"  name        = \"A\"\n" +
		"  description = \"d\"\n" +
		"  runs {\n" +
		"    using = \"composite\"\n" +
		"    steps = [step.hi]\n" +
		"  }\n" +
		"}\n"
}
