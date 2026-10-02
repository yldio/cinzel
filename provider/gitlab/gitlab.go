// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yldio/cinzel/internal/fsutil"
	"github.com/yldio/cinzel/provider"
)

const (
	defaultParseOutputDirectory   = "."
	defaultUnparseOutputDirectory = "./cinzel"

	providerName = "gitlab"
	providerDesc = "GitLab CI/CD Pipelines https://about.gitlab.com/stages-devops-lifecycle/continuous-integration/"
	parseDesc    = "Convert HCL definitions to GitLab CI/CD YAML"
	unparseDesc  = "Convert GitLab CI/CD YAML to HCL definitions"
)

// GitLab is the provider implementation for GitLab CI/CD pipelines.
type GitLab struct{}

// New returns a new GitLab provider instance.
func New() *GitLab { return &GitLab{} }

// GetProviderName returns the provider identifier.
func (p *GitLab) GetProviderName() string { return providerName }

// GetDescription returns a human-readable provider description.
func (p *GitLab) GetDescription() string { return providerDesc }

// GetParseDescription returns a parse operation description.
func (p *GitLab) GetParseDescription() string { return parseDesc }

// GetUnparseDescription returns an unparse operation description.
func (p *GitLab) GetUnparseDescription() string { return unparseDesc }

// Parse converts HCL pipeline definitions to GitLab YAML.
func (p *GitLab) Parse(opts provider.ProviderOps) error {
	inputPath, err := resolveInputPath(opts)
	if err != nil {
		return err
	}

	body, sources, err := fsutil.ParseHCLInput(inputPath, opts.Recursive)
	if err != nil {
		return err
	}

	pipeline, comments, err := parseHCLToPipeline(body, sources)
	if err != nil {
		return err
	}

	// Nothing was declared. Writing here would leave a "{}" document behind.
	if len(pipeline) == 0 {
		return errNoDefinitions
	}

	outputBytes, err := marshalPipelineYAML(pipeline, comments)
	if err != nil {
		return err
	}

	// No source is recorded: a pipeline is one file built from every HCL file
	// read, so there is no single file to name, and nothing here prunes.
	outputBytes = fsutil.PrependGeneratedMarker(outputBytes, providerName, "")

	outputPath := filepath.Join(resolveParseOutputDirectory(opts), ".gitlab-ci.yml")

	if opts.DryRun {
		fmt.Printf("# file: %s\n", outputPath)
		fmt.Println(string(outputBytes))

		return nil
	}

	if err := fsutil.WriteFile(outputPath, outputBytes); err != nil {
		return err
	}

	return nil
}

// Unparse converts GitLab YAML to HCL pipeline definitions.
func (p *GitLab) Unparse(opts provider.ProviderOps) error {
	inputPath, err := resolveInputPath(opts)
	if err != nil {
		return err
	}

	files, err := fsutil.ListFilesWithExtensions(inputPath, opts.Recursive, ".yaml", ".yml")
	if err != nil {
		return err
	}

	// Nothing was read. Reporting success here says the input was converted
	// when it was not: pointing at the wrong directory, or forgetting
	// --recursive with the pipeline a level down, both land exactly here.
	if len(files) == 0 {
		return errNoYAMLFiles
	}

	outputDir := resolveUnparseOutputDirectory(opts)

	// The output name comes from the input's basename, so a recursive run over
	// two directories each holding a ".gitlab-ci.yml" aimed both at one path.
	takenNames := make(map[string]struct{}, len(files))

	// Every file is converted before any of them is written. Returning at the
	// first failure left the files converted before it on disk and never read
	// the ones after, so which half of a directory survived was decided by
	// where the failing file sorted.
	converted := make([]unparsedFile, 0, len(files))

	// Collected rather than returned at the first, so a directory of broken
	// pipelines is read through in one run instead of one key at a time.
	failures := make([]error, 0)

	for _, file := range files {
		yamlBytes, err := os.ReadFile(file)
		if err != nil {
			// Named, like the two below. A directory run that returned this
			// bare said a file could not be read without saying which one.
			failures = append(failures, fmt.Errorf("error in file '%s': %w", file, err))

			continue
		}

		// Named, the way the conversion error below is. A directory run reads
		// every ".yaml" in the tree, and the parser's complaint carries a line
		// and column but no file: on its own it says a pipeline somewhere
		// failed to parse without saying which.
		doc, err := parseYAMLDocument(yamlBytes)
		if err != nil {
			failures = append(failures, fmt.Errorf("error in file '%s': %w", file, err))

			continue
		}

		if !classifyPipelineDocument(doc) {
			// Said nothing before. A file skipped in the middle of a directory
			// run left no trace at all, so pointing unparse at a directory
			// holding one pipeline and four other YAML files reported the same
			// success as a run that converted every one of them. The GitHub
			// provider has warned here for the same reason.
			//
			// Skipping is not failing: cinzel never had this file, so passing
			// over it loses nothing and the run goes on.
			warnf("skipping '%s': not a pipeline", file)

			continue
		}

		baseName := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		hclBytes, err := pipelineToHCL(doc, baseName, documentComments(yamlBytes))
		if err != nil {
			failures = append(failures, fmt.Errorf("error in file '%s': %w", file, err))

			continue
		}

		converted = append(converted, unparsedFile{name: baseName, hcl: hclBytes})
	}

	// Before the no-definitions check below. "found" used to be set before the
	// conversion rather than after it, so a run whose only pipeline failed had
	// already recorded that it found one; checking the failures first makes
	// that ordering stop mattering, and reports what the user can act on.
	if len(failures) > 0 {
		return errors.Join(failures...)
	}

	// Files were read but none of them held a pipeline, so again nothing was
	// written. The same silence hides the same mistake. A dry run counts: it
	// found the pipeline and only skipped the write it was told to skip.
	if len(converted) == 0 {
		return errNoDefinitions
	}

	for _, out := range converted {
		outputPath := filepath.Join(outputDir, fsutil.UniqueOutputName(takenNames, out.name)+".hcl")

		if opts.DryRun {
			fmt.Printf("# file: %s\n", outputPath)
			fmt.Println(string(out.hcl))

			continue
		}

		if err := fsutil.WriteFile(outputPath, out.hcl); err != nil {
			return err
		}
	}

	return nil
}

// unparsedFile is one converted pipeline waiting to be written: the output
// basename it was given, and the HCL it became.
type unparsedFile struct {
	name string
	hcl  []byte
}
