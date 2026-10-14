// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yldio/cinzel/provider"
)

func resolveInputPath(opts provider.ProviderOps) (string, error) {
	paths, err := resolveInputPaths(opts)
	if err != nil {
		return "", err
	}

	if len(paths) > 1 {
		return "", errOneInputPathOnly
	}

	return paths[0], nil
}

// resolveInputPaths is every directory a run reads, or the single file it was
// given, each already under the run's root.
//
// Several directories because a repository can keep composite actions in
// folders of its own beside the workflows, and converting them in one run is
// what lets the HCL reference across them.
func resolveInputPaths(opts provider.ProviderOps) ([]string, error) {
	dirs := opts.InputDirectories()

	if opts.File == "" && len(dirs) == 0 {
		return nil, errInputPathRequired
	}

	if opts.File != "" && len(dirs) > 0 {
		return nil, errInputPathConflict
	}

	if opts.File != "" {
		stat, err := os.Stat(opts.UnderRoot(opts.File))
		if err != nil {
			return nil, err
		}

		// --file names a file. It reached a directory only because the path
		// goes on to be stat'd again further down, and a reader following the
		// flag's own wording had no reason to expect it: --directory is the
		// flag for that, the two refuse to be set together, and nothing said
		// one stood in for the other.
		if stat.IsDir() {
			return nil, fmt.Errorf("%w: %s", errFileIsADirectory, opts.File)
		}

		return []string{opts.UnderRoot(opts.File)}, nil
	}

	if err := checkDirectoriesDoNotOverlap(dirs, opts); err != nil {
		return nil, err
	}

	rooted := make([]string, 0, len(dirs))

	for _, dir := range dirs {
		rooted = append(rooted, opts.UnderRoot(dir))
	}

	return rooted, nil
}

// checkDirectoriesDoNotOverlap refuses a run where one directory already
// contains another.
//
// Only under --recursive. Without it "-d a" reads a's own files and nothing
// below, so "-d a -d a/b" names two sets that do not meet and both are
// meaningful. With it the first already covers the second, and the files
// underneath would be read twice: converted twice, written twice, and the
// second copy refused for sharing a name with the first.
func checkDirectoriesDoNotOverlap(dirs []string, opts provider.ProviderOps) error {
	if !opts.Recursive || len(dirs) < 2 {
		return nil
	}

	for i, outer := range dirs {
		for j, inner := range dirs {
			if i == j {
				continue
			}

			contains, err := directoryContains(opts.UnderRoot(outer), opts.UnderRoot(inner))
			if err != nil {
				return err
			}

			if contains {
				return fmt.Errorf("%w: '%s' already includes '%s'", errOverlappingDirectories, outer, inner)
			}
		}
	}

	return nil
}

// directoryContains reports whether a recursive read of outer reaches inner.
func directoryContains(outer string, inner string) (bool, error) {
	absOuter, err := filepath.Abs(outer)
	if err != nil {
		return false, err
	}

	absInner, err := filepath.Abs(inner)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(absInner, absOuter+string(filepath.Separator)), nil
}

func resolveParseOutputDirectory(opts provider.ProviderOps) string {
	if opts.OutputDirectory != "" {
		return opts.OutputDirectory
	}

	return opts.UnderRoot(defaultParseOutputDirectory)
}

func resolveUnparseOutputDirectory(opts provider.ProviderOps) string {
	if opts.OutputDirectory != "" {
		return opts.OutputDirectory
	}

	return opts.UnderRoot(defaultUnparseOutputDirectory)
}
