// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/internal/fsutil"
)

func (cmd *Cli) fmtCommand() *cli.Command {
	return &cli.Command{
		Name:  "fmt",
		Usage: "Rewrite HCL files in the canonical layout",
		Action: func(ctx context.Context, c *cli.Command) error {
			paths, err := fmtPaths(c)
			if err != nil {
				return err
			}

			files := make([]string, 0, len(paths))

			for _, path := range paths {
				found, err := fsutil.ListFilesWithExtensions(path, c.Bool("recursive"), ".hcl")
				if err != nil {
					// A path that is not there was typed, or written into
					// the configuration, by whoever is running this. Without
					// the mark it is reported as cinzel's own fault, under
					// an invitation to open an issue about a missing
					// directory.
					return cinzelerror.UserInput(err)
				}

				files = append(files, found...)
			}

			// A run that formatted nothing is a run that was pointed
			// somewhere wrong. Reporting "0 checked" and exiting 0 passes a
			// --check step while checking no file at all, which is the one
			// answer that must not come back clean.
			if len(files) == 0 {
				return fmt.Errorf("%w: %s", errNoHCLFiles, strings.Join(paths, ", "))
			}

			return cmd.formatFiles(files, c.Bool("check"))
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "Format a single HCL `FILE`",
			},
			&cli.StringFlag{
				Name:    "directory",
				Aliases: []string{"d"},
				Value:   "",
				Usage:   "Format the HCL files in `DIRECTORY` (default: cinzel)",
			},
			&cli.BoolFlag{
				Name:    "recursive",
				Aliases: []string{"r"},
				Value:   false,
				Usage:   "Reads the directory recursively",
			},
			&cli.BoolFlag{
				Name:  "check",
				Value: false,
				Usage: "Name the files that are not formatted and exit non-zero, writing nothing",
			},
		},
	}
}

// fmtPaths works out what to format: what the flags name, or failing that
// what the configuration file says the HCL is, or failing that "cinzel".
//
// The configuration is read because the directory is already written down
// there, under parse as the HCL that is read and under unparse as where the
// HCL is written. A repository keeping its HCL somewhere else said so once,
// and fmt reading only its own default formatted nothing and said so in a
// line that looked like success.
func fmtPaths(c *cli.Command) ([]string, error) {
	if path := c.String("file"); path != "" {
		if err := validateRelativePath(path); err != nil {
			return nil, fmt.Errorf("--file: %w", err)
		}

		if err := requireFileNotDirectory(path); err != nil {
			return nil, err
		}

		return []string{path}, nil
	}

	if path := c.String("directory"); path != "" {
		if err := validateRelativePath(path); err != nil {
			return nil, fmt.Errorf("--directory: %w", err)
		}

		return []string{path}, nil
	}

	paths, err := configuredHCLPaths()
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return []string{"cinzel"}, nil
	}

	return paths, nil
}

// configuredHCLPaths returns every path in .cinzelrc.yaml that holds HCL,
// deduplicated and in a stable order.
//
// Which keys those are follows the direction: unparse writes HCL, so its
// output-directory is HCL, and parse reads HCL, so its file and directory
// are. The warnings the loader raises belong to the command whose settings
// they are about, and are left for parse and unparse to report.
func configuredHCLPaths() ([]string, error) {
	seen := map[string]struct{}{}
	paths := make([]string, 0, 4)

	add := func(path string) {
		if path == "" {
			return
		}

		if _, ok := seen[path]; ok {
			return
		}

		seen[path] = struct{}{}
		paths = append(paths, path)
	}

	for _, providerName := range []string{"github", "gitlab"} {
		unparse, _, err := loadProviderCommandConfig(configFilename, providerName, "unparse")
		if err != nil {
			return nil, err
		}

		add(unparse.outputDirectory)

		parse, _, err := loadProviderCommandConfig(configFilename, providerName, "parse")
		if err != nil {
			return nil, err
		}

		add(parse.directory)
		add(parse.file)
	}

	sort.Strings(paths)

	return paths, nil
}

// formatFiles rewrites each file, or under check only names the ones that
// would change.
//
// A file is parsed before it is formatted. hclwrite.Format works on tokens
// and not on a parsed document, so it reformats a file it cannot parse as
// readily as one it can, and writing that back leaves broken HCL laid out
// neatly enough to look deliberate.
func (cmd *Cli) formatFiles(files []string, check bool) error {
	changed := 0

	for _, path := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if _, diags := hclsyntax.ParseConfig(content, path, hcl.Pos{Line: 1, Column: 1}); diags.HasErrors() {
			return fmt.Errorf("%s: %w", path, cinzelerror.ProcessHCLDiags(diags))
		}

		formatted := hclwrite.Format(content)

		if fsutil.SameContent(content, formatted) {
			continue
		}

		changed++

		_, _ = fmt.Fprintln(cmd.Writer, path)

		if check {
			continue
		}

		if err := fsutil.WriteFile(path, formatted); err != nil {
			return err
		}
	}

	_, _ = fmt.Fprintf(cmd.Writer, "\n%d checked, %d %s\n", len(files), changed, formattedVerb(check))

	if check && changed > 0 {
		return fmt.Errorf("%w: %d of %d", errUnformatted, changed, len(files))
	}

	return nil
}

func formattedVerb(check bool) string {
	if check {
		return "not formatted"
	}

	return "formatted"
}
