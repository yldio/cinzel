// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import "github.com/yldio/cinzel/provider"

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
