// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yldio/cinzel/internal/cinzelerror"
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
func (p *captureProvider) GetProviderName() string       { return "github" }
func (p *captureProvider) GetDescription() string        { return "github" }
func (p *captureProvider) GetParseDescription() string   { return "parse" }
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

		want := "# file: " + filepath.Join(".github", "workflows", "ci.yaml")

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

		want := "# file: " + filepath.Join("custom", "ci.yaml")

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

		want := "# file: " + filepath.Join(".github", "workflows", "ci.yaml")

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

// single-file and filename were type-checked and then thrown away: nothing
// read them. They are gone, so they warn like any other unknown key.
func TestDroppedKeysWarn(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    single-file: true\n    filename: ci\n"))

		app, errBuf, p := newConfigTestApp(t)
		err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		got := errBuf.String()
		want := "warning: .cinzelrc.yaml.github.parse.filename: unknown key\n" +
			"warning: .cinzelrc.yaml.github.parse.single-file: unknown key\n"

		if got != want {
			t.Fatalf("warnings = %q, want %q", got, want)
		}
	})
}

// The configuration file is committed and read on every machine that checks
// the repo out, so a path in it that names one machine's disk is a path that
// breaks for everyone else.
func TestUnportablePathsInConfigFail(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{"absolute file", "github:\n  parse:\n    file: " + absoluteTestPath() + "\n"},
		{"absolute directory", "github:\n  parse:\n    directory: " + absoluteTestPath() + "\n"},
		{"absolute output-directory", "github:\n  parse:\n    output-directory: " + absoluteTestPath() + "\n"},
		{"home-relative file", "github:\n  parse:\n    file: ~/cinzel/main.hcl\n"},
		{"home-relative output-directory", "github:\n  parse:\n    output-directory: ~/out\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeFile(t, configFilename, []byte(tc.yaml))

				app, _, p := newConfigTestApp(t)
				err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})

				if err == nil {
					t.Fatalf("Execute() error = nil, want error")
				}

				if !strings.Contains(err.Error(), "must be a relative path") {
					t.Fatalf("error = %v, want it to name the rule", err)
				}

				if p.parseOpts.File != "" || p.parseOpts.Directory != "" || p.parseOpts.OutputDirectory != "" {
					t.Fatalf("opts should be zero on config error, got %+v", p.parseOpts)
				}
			})
		})
	}
}

// The configuration file is committed, so a path in it is a path every reader
// runs without having written it. One that climbs out of the checkout writes
// generated YAML somewhere else, and parse prunes what it finds there.
func TestEscapingPathsInConfigFail(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{"output-directory", "github:\n  parse:\n    output-directory: ../../tmp/evil\n"},
		{"directory", "github:\n  parse:\n    directory: ../elsewhere\n"},
		{"file", "github:\n  parse:\n    file: ../../main.hcl\n"},
		{"only after cleaning", "github:\n  parse:\n    output-directory: a/../../out\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeFile(t, configFilename, []byte(tc.yaml))

				app, _, p := newConfigTestApp(t)
				err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})

				if err == nil {
					t.Fatalf("Execute() error = nil, want error")
				}

				if !strings.Contains(err.Error(), "must stay inside") {
					t.Fatalf("error = %v, want it to name the rule", err)
				}

				if p.parseOpts.File != "" || p.parseOpts.Directory != "" || p.parseOpts.OutputDirectory != "" {
					t.Fatalf("opts should be zero on config error, got %+v", p.parseOpts)
				}
			})
		})
	}
}

// A ".." that names nothing is an ordinary directory name, not an escape.
func TestDotPrefixedPathsInConfigAreKept(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    output-directory: ..hidden/out\n"))

		app, _, p := newConfigTestApp(t)
		if err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p}); err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}

		if want := filepath.Join("..hidden", "out"); p.parseOpts.OutputDirectory != want {
			t.Fatalf("OutputDirectory = %q, want %q", p.parseOpts.OutputDirectory, want)
		}
	})
}

// One committed spelling has to work on every OS, so the forward slashes a
// config is written with become whatever the running one separates with.
func TestConfigPathSeparatorsAreNormalized(t *testing.T) {
	withTempWorkingDir(t, func() {
		writeFile(t, configFilename, []byte("github:\n  parse:\n    file: ./cinzel/main.hcl\n"))

		app, _, p := newConfigTestApp(t)

		if err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		if want := filepath.FromSlash("./cinzel/main.hcl"); p.parseOpts.File != want {
			t.Fatalf("parse file = %q, want %q", p.parseOpts.File, want)
		}
	})
}

// absoluteTestPath returns a path the running OS reads as absolute, which is
// spelled differently on Windows.
func absoluteTestPath() string {
	if filepath.IsAbs("/tmp/cinzel") {
		return "/tmp/cinzel"
	}

	return `C:\cinzel`
}

// TestConfigErrorsAreTheAuthorsToFix walks every refusal the configuration file
// can raise about its own contents. Each names a key or a value someone wrote in
// a file they keep, so the open-an-issue line New adds by default sent them to
// the issue tracker over their own typo.
func TestConfigErrorsAreTheAuthorsToFix(t *testing.T) {
	for _, tc := range []struct {
		name string
		yaml string
	}{
		{"not YAML at all", "github:\n  parse:\n   - x\n  : y\n"},
		{"root is not a mapping", "- a\n"},
		{"provider is not a mapping", "github: nope\n"},
		{"command is not a mapping", "github:\n  parse: nope\n"},
		{"path is not a string", "github:\n  parse:\n    file: 3\n"},
		{"yml is not a boolean", "github:\n  parse:\n    yml: \"yes\"\n"},
		{"file and directory together", "github:\n  parse:\n    file: a.hcl\n    directory: d\n"},
		{"path names one machine", "github:\n  parse:\n    output-directory: /tmp/elsewhere\n"},
		{"path climbs out", "github:\n  parse:\n    output-directory: ../elsewhere\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withTempWorkingDir(t, func() {
				writeFile(t, configFilename, []byte(tc.yaml))

				app, _, p := newConfigTestApp(t)
				err := app.Execute([]string{"cinzel", "github", "parse"}, []provider.Provider{p})

				if err == nil {
					t.Fatalf("Execute() error = nil, want error")
				}

				if !cinzelerror.IsUserInput(err) {
					t.Fatalf("error = %v, which asks the author to open an issue about their own file", err)
				}
			})
		})
	}
}
