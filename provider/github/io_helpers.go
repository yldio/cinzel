// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yldio/cinzel/provider"
	"github.com/yldio/cinzel/provider/github/step"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	yamlv3 "gopkg.in/yaml.v3"
)

func resolveInputPath(opts provider.ProviderOps) (string, error) {
	if opts.File == "" && opts.Directory == "" {
		return "", errInputPathRequired
	}

	if opts.File != "" && opts.Directory != "" {
		return "", errInputPathConflict
	}

	if opts.File != "" {
		return opts.File, nil
	}

	return opts.Directory, nil
}

func resolveParseOutputDirectory(opts provider.ProviderOps) string {
	if opts.OutputDirectory != "" {
		return opts.OutputDirectory
	}

	return defaultParseOutputDirectory
}

func resolveUnparseOutputDirectory(opts provider.ProviderOps) string {
	if opts.OutputDirectory != "" {
		return opts.OutputDirectory
	}

	return defaultUnparseOutputDirectory
}

func resolveParseFilename(opts provider.ProviderOps) string {
	ext := workflowExt(opts)

	if opts.File == "" {
		return "steps" + ext
	}

	name := strings.TrimSuffix(filepath.Base(opts.File), filepath.Ext(opts.File))

	if name == "" {
		return "steps" + ext
	}

	return name + ext
}

// workflowExt returns the YAML file extension for workflow output files.
// Defaults to ".yaml"; returns ".yml" when opts.YML is true.
func workflowExt(opts provider.ProviderOps) string {
	if opts.YML {
		return ".yml"
	}

	return ".yaml"
}

// keepWholeNumbersExactInYAML runs the retagging pass over the bytes rather
// than over a node the caller already holds.
//
// parseYAMLDocument reads the file with yaml.v3 and retags a run of digits too
// long for an integer, so the digits survive as text. The step-only path below
// reads the same file again through a second reader, which has no such pass:
// cty resolves the run to a number it cannot hold every digit of, so a 180
// digit ID came back with its tail replaced by zeros, at exit 0. Running the
// pass over the bytes first puts the two readers back on the same document.
func keepWholeNumbersExactInYAML(content []byte) ([]byte, error) {
	var node yamlv3.Node

	if err := yamlv3.Unmarshal(content, &node); err != nil {
		return nil, err
	}

	keepWholeNumbersExact(&node)

	return yamlv3.Marshal(&node)
}

func parseStepsFromYAML(content []byte) ([]step.Step, error) {
	content, err := keepWholeNumbersExactInYAML(content)
	if err != nil {
		return nil, err
	}

	// The document is read as a whole rather than as a map of a single element
	// type: a map forces every step to unify to one type, so two steps that do
	// not carry exactly the same keys would be rejected outright.
	val, err := ctyyaml.Unmarshal(content, cty.DynamicPseudoType)
	if err != nil {
		return nil, err
	}

	if val.IsNull() || !val.IsKnown() {
		return nil, nil
	}

	if !val.Type().IsObjectType() && !val.Type().IsMapType() {
		return nil, fmt.Errorf("expected top-level map, found %s", val.Type().FriendlyName())
	}

	rawMap := val.AsValueMap()

	// The step-only path is the end of the detection chain, so every document
	// that is neither a workflow nor an action arrives here — a dependabot
	// config and an issue template among them. A step is a mapping, so a
	// document holding anything else at the top is not a set of steps. Handing
	// one to the decoder anyway came back "not a valid type", which aborted the
	// whole directory it sat in before the workflows beside it were reached.
	for _, v := range rawMap {
		if v.IsNull() || !v.IsKnown() || !v.Type().IsObjectType() {
			return nil, nil
		}
	}

	ids := make([]string, 0, len(rawMap))

	for id := range rawMap {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	steps := make([]step.Step, 0, len(ids))

	for _, id := range ids {
		var s step.Step

		if err := s.PreDecode(rawMap[id]); err != nil {
			return nil, err
		}

		s.Update(id)
		steps = append(steps, s)
	}

	return steps, nil
}

// checkFilenameStaysInside refuses a filename that would place the output
// somewhere other than the directory it was asked for. A filename is written
// straight into the output path, so "../../x" walked out of it and an absolute
// path ignored it altogether, both without a word. That turns a pipeline
// definition into a write anywhere the process can reach.
//
// A plain subdirectory is still allowed: it is the only way an action can sit
// under its own folder, which is what the action writer already relies on.
func checkFilenameStaysInside(filename string) error {
	if filename == "" {
		return nil
	}

	// A filename is written in HCL, and the same HCL is read on every
	// platform, so what counts as a separator or as rooted cannot be left to
	// the one running. Both separators are folded to "/" and judged there:
	// "/x" is not absolute on Windows, "C:x" is not absolute anywhere, and
	// "..\\x" is a single name on Linux, yet none of them belongs under the
	// output directory.
	slashed := strings.ReplaceAll(filename, `\`, "/")

	if filepath.IsAbs(filename) || strings.HasPrefix(slashed, "/") || hasDriveLetter(filename) {
		return fmt.Errorf("%w: %s", errFilenameEscapes, filename)
	}

	if clean := path.Clean(slashed); clean == ".." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("%w: %s", errFilenameEscapes, filename)
	}

	return nil
}

// checkOutputPaths refuses two definitions that write one file. It is the
// filename guards' counterpart at the point the extension is known: an action
// is written to "<filename>/action.yml" and a workflow to "<filename>.yml", so
// a workflow called "build/action" lands exactly where the action called
// "build" does. Each kind compared its own filenames against its own kind, and
// compared the filenames rather than the paths they become, so neither saw the
// other and the second write landed on the first.
//
// Run before anything is written, so a collision leaves the output directory as
// it was. The comparison folds case, for the reason claimFilename gives.
func checkOutputPaths(workflows []WorkflowYAMLFile, actions []ActionYAMLFile, outputDir, ext string) error {
	taken := make(map[string]string, len(workflows)+len(actions))

	claim := func(path, filename string) error {
		key := strings.ToLower(filepath.Clean(path))

		if other, ok := taken[key]; ok {
			return fmt.Errorf("%w: '%s' and '%s' both write to '%s'", errDuplicateFilename, other, filename, path)
		}

		taken[key] = filename

		return nil
	}

	for _, w := range workflows {
		if err := claim(filepath.Join(outputDir, w.Filename+ext), w.Filename); err != nil {
			return err
		}
	}

	for _, a := range actions {
		if err := claim(filepath.Join(outputDir, a.Filename, "action.yml"), a.Filename); err != nil {
			return err
		}
	}

	return nil
}

// hasDriveLetter reports whether the name starts with a Windows drive, such
// as "C:" or "C:x". filepath.VolumeName answers this only when the tool is
// running on Windows, and the same HCL is read everywhere.
func hasDriveLetter(name string) bool {
	if len(name) < 2 || name[1] != ':' {
		return false
	}

	c := name[0]

	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// claimFilename records a filename and refuses one already taken. Two
// definitions naming the same file wrote one on top of the other and said
// nothing: the output held the last one and the rest were simply gone.
//
// The comparison folds case, because a filename that differs only in case is
// the same file on macOS and Windows, and the same input then produces
// different output depending on where it ran.
func claimFilename(taken map[string]string, filename, id string) error {
	key := strings.ToLower(filepath.Clean(filepath.FromSlash(filename)))

	if other, ok := taken[key]; ok {
		return fmt.Errorf("%w: '%s' and '%s' both write to '%s'", errDuplicateFilename, other, id, filename)
	}

	taken[key] = id

	return nil
}
