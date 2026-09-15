// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yldio/cinzel/provider"
	"github.com/yldio/cinzel/provider/github/step"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
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

func parseStepsFromYAML(content []byte) ([]step.Step, error) {
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

	// VolumeName catches a Windows drive or share, which IsAbs misses for a
	// drive-relative path such as "C:x", and is empty everywhere else.
	if filepath.IsAbs(filename) || filepath.VolumeName(filename) != "" {
		return fmt.Errorf("%w: %s", errFilenameEscapes, filename)
	}

	clean := filepath.Clean(filepath.FromSlash(filename))

	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("%w: %s", errFilenameEscapes, filename)
	}

	return nil
}
