// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package github

import (
	"fmt"
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
	if opts.File == "" {
		return "steps.yaml"
	}

	name := strings.TrimSuffix(filepath.Base(opts.File), filepath.Ext(opts.File))

	if name == "" {
		return "steps.yaml"
	}

	return name + ".yaml"
}

func parseStepsFromYAML(content []byte) ([]step.Step, error) {
	typ := cty.Map(cty.DynamicPseudoType)
	val, err := ctyyaml.Unmarshal(content, typ)
	if err != nil {
		return nil, err
	}

	if val.IsNull() || !val.IsKnown() {
		return nil, nil
	}

	if !val.Type().IsMapType() {
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
