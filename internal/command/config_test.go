// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package command

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/provider"
	githubprovider "github.com/yldio/cinzel/provider/github"
)

type captureProvider struct {
	parseOpts   provider.ProviderOps
	unparseOpts provider.ProviderOps
}

func (p *captureProvider) Parse(opts provider.ProviderOps) error {
	p.parseOpts = opts

	return nil
}

func (p *captureProvider) Unparse(opts provider.ProviderOps) error {
	p.unparseOpts = opts

	return nil
}

func (p *captureProvider) GetProviderName() string { return "github" }

func (p *captureProvider) GetDescription() string { return "github" }

func (p *captureProvider) GetParseDescription() string { return "parse" }

func (p *captureProvider) GetUnparseDescription() string { return "unparse" }

func TestConfigSetsParseOutputDirectory(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: .github/workflows\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.OutputDirectory != ".github/workflows" {
			t.Fatalf("parse output-directory = %q, want %q", p.parseOpts.OutputDirectory, ".github/workflows")
		}
	})
}

func TestConfigSetsParseDirectory(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    directory: ./cinzel\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.Directory != "./cinzel" {
			t.Fatalf("parse directory = %q, want %q", p.parseOpts.Directory, "./cinzel")
		}

		if p.parseOpts.File != "" {
			t.Fatalf("parse file = %q, want empty string", p.parseOpts.File)
		}
	})
}

func TestConfigSetsUnparseOutputDirectory(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  unparse:\n    output-directory: ./cinzel\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "unparse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.unparseOpts.OutputDirectory != "./cinzel" {
			t.Fatalf("unparse output-directory = %q, want %q", p.unparseOpts.OutputDirectory, "./cinzel")
		}
	})
}

func TestCLIOutputDirectoryOverridesConfig(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: .github/workflows\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse", "--output-directory", "./custom"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.OutputDirectory != "./custom" {
			t.Fatalf("parse output-directory = %q, want %q", p.parseOpts.OutputDirectory, "./custom")
		}
	})
}

func TestEmptyCLIOutputDirectoryOverridesConfig(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: .github/workflows\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse", "--output-directory", ""}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.OutputDirectory != "" {
			t.Fatalf("parse output-directory = %q, want empty string", p.parseOpts.OutputDirectory)
		}
	})
}

func TestMissingConfigDoesNotFail(t *testing.T) {
	withTempWorkingDir(t, func() {
		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.OutputDirectory != "" {
			t.Fatalf("parse output-directory = %q, want empty string", p.parseOpts.OutputDirectory)
		}
	})
}

func TestInvalidActiveProviderFieldTypeFails(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: 123\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err == nil {
			t.Fatalf("Execute() error = nil, want error")
		}

		if p.parseOpts.OutputDirectory != "" {
			t.Fatalf("parse output-directory = %q, want empty string", p.parseOpts.OutputDirectory)
		}
	})
}

func TestConfigFileAndDirectoryConflictFails(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    file: ./cinzel/workflow.hcl\n    directory: ./cinzel\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err == nil {
			t.Fatalf("Execute() error = nil, want error")
		}

		if p.parseOpts.File != "" || p.parseOpts.Directory != "" {
			t.Fatalf("parse opts should be zero on config error, got file=%q directory=%q", p.parseOpts.File, p.parseOpts.Directory)
		}
	})
}

func TestCLIFileOverridesConfigDirectory(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    directory: ./cinzel\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse", "--file", "./one.hcl"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.File != "./one.hcl" {
			t.Fatalf("parse file = %q, want %q", p.parseOpts.File, "./one.hcl")
		}

		if p.parseOpts.Directory != "" {
			t.Fatalf("parse directory = %q, want empty string", p.parseOpts.Directory)
		}
	})
}

func TestInvalidInactiveProviderFieldTypeDoesNotFail(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: .github/workflows\ngitlab:\n  parse:\n    output-directory: 123\n"))

		app, _, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if p.parseOpts.OutputDirectory != ".github/workflows" {
			t.Fatalf("parse output-directory = %q, want %q", p.parseOpts.OutputDirectory, ".github/workflows")
		}
	})
}

func TestUnknownKeysEmitDeterministicWarnings(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  z-last: true\n  parse:\n    x-second: true\n    output-directory: .github/workflows\n    a-first: true\n"))

		app, errBuf, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := errBuf.String()
		want := "warning: .cinzelrc.yaml.github.parse.a-first: unknown key\n" +
			"warning: .cinzelrc.yaml.github.parse.x-second: unknown key\n" +
			"warning: .cinzelrc.yaml.github.z-last: unknown key\n"

		if got != want {
			t.Fatalf("warnings = %q, want %q", got, want)
		}
	})
}

func TestDryRunUsesConfigResolvedOutputDirectory(t *testing.T) {
	withTempWorkingDir(t, func() {
		inputFile := "workflow.hcl"
		writeFile(t, inputFile, []byte(`
step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "ci"

  on "push" {}

  jobs = [job.build]
}
`))

		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: .github/workflows\n"))

		out := new(bytes.Buffer)
		app := New(out, "v.test")

		stdout := captureStdout(t, func() {
			err := app.Execute([]string{"cinzel", "github", "parse", "--file", inputFile, "--dry-run"}, []provider.Provider{githubprovider.New()})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})

		want := "# file: .github/workflows/ci.yaml"
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want to contain %q", stdout, want)
		}
	})
}

func TestDryRunCLIOutputDirectoryOverridesConfigPath(t *testing.T) {
	withTempWorkingDir(t, func() {
		inputFile := "workflow.hcl"
		writeFile(t, inputFile, []byte(`
step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "ci"

  on "push" {}

  jobs = [job.build]
}
`))

		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: .github/workflows\n"))

		out := new(bytes.Buffer)
		app := New(out, "v.test")

		stdout := captureStdout(t, func() {
			err := app.Execute([]string{"cinzel", "github", "parse", "--file", inputFile, "--dry-run", "--output-directory", "./custom"}, []provider.Provider{githubprovider.New()})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})

		want := "# file: custom/ci.yaml"
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want to contain %q", stdout, want)
		}
	})
}

func TestDryRunUsesConfigDirectoryWithoutCLIInput(t *testing.T) {
	withTempWorkingDir(t, func() {
		inputDir := "cinzel"
		if err := os.Mkdir(inputDir, 0o755); err != nil {
			t.Fatalf("Mkdir(%q) error = %v", inputDir, err)
		}

		writeFile(t, filepath.Join(inputDir, "workflow.hcl"), []byte(`
step "echo" {
  run = "echo hi"
}

job "build" {
  runs_on {
    runners = "ubuntu-latest"
  }

  steps = [step.echo]
}

workflow "ci" {
  filename = "ci"

  on "push" {}

  jobs = [job.build]
}
`))

		writeFile(t, configFilename, []byte("github:\n  parse:\n    directory: ./cinzel\n    output-directory: .github/workflows\n"))

		out := new(bytes.Buffer)
		app := New(out, "v.test")

		stdout := captureStdout(t, func() {
			err := app.Execute([]string{"cinzel", "github", "parse", "--dry-run"}, []provider.Provider{githubprovider.New()})
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
		})

		want := "# file: .github/workflows/ci.yaml"
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want to contain %q", stdout, want)
		}
	})
}

func newConfigTestApp(t *testing.T) (*Cli, *bytes.Buffer, *captureProvider) {
	t.Helper()

	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	app := New(out, "v.test")
	app.Cmd.ErrWriter = errOut

	return app, errOut, &captureProvider{}
}

func withTempWorkingDir(t *testing.T, fn func()) {
	t.Helper()

	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Chdir(%q) error = %v", tempDir, err)
	}

	t.Cleanup(func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Fatalf("Chdir(%q) restore error = %v", originalDir, chdirErr)
		}
	})

	fn()
}

func writeFile(t *testing.T, name string, content []byte) {
	t.Helper()

	if err := os.WriteFile(filepath.Clean(name), content, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", name, err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	originalStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		os.Stdout = originalStdout
		_ = r.Close()
		_ = w.Close()
	})

	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("stdout close error = %v", err)
	}

	os.Stdout = originalStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stdout error = %v", err)
	}

	if closeErr := r.Close(); closeErr != nil {
		t.Fatalf("stdout read pipe close error = %v", closeErr)
	}

	return string(out)
}
