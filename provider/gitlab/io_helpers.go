// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"fmt"
	"os"

	"github.com/yldio/cinzel/provider"
)

func resolveInputPath(opts provider.ProviderOps) (string, error) {
	if opts.File == "" && opts.Directory == "" {
		return "", errInputPathRequired
	}

	if opts.File != "" && opts.Directory != "" {
		return "", errInputPathConflict
	}

	if opts.File != "" {
		stat, err := os.Stat(opts.File)
		if err != nil {
			return "", err
		}

		// --file names a file. It reached a directory only because the path
		// goes on to be stat'd again further down, and a reader following the
		// flag's own wording had no reason to expect it: --directory is the
		// flag for that, the two refuse to be set together, and nothing said
		// one stood in for the other.
		if stat.IsDir() {
			return "", fmt.Errorf("%w: %s", errFileIsADirectory, opts.File)
		}

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
