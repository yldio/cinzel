// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"io"
	"net/mail"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/cinzelerror"
	"github.com/yldio/cinzel/provider"
)

const (
	appName   = "cinzel"
	copyright = "(c) 2024-2026 YLD Limited"
)

// Cli holds the CLI application state including the output writer and root command.
type Cli struct {
	Writer io.Writer
	Cmd    *cli.Command
}

// Execute registers the given providers and runs the CLI with the supplied arguments.
func (cmd *Cli) Execute(osArgs []string, providers []provider.Provider) error {
	for _, p := range providers {
		ap := cmd.addProvider(p)
		cmd.Cmd.Commands = append(cmd.Cmd.Commands, ap)
	}

	cmd.Cmd.Commands = append(cmd.Cmd.Commands, cmd.initCommand())

	if err := cmd.Cmd.Run(context.Background(), osArgs); err != nil {
		_, _ = fmt.Fprintf(cmd.Writer, "%s\n", cinzelerror.New(err).Err.Error())

		return err
	}

	return nil
}

// New creates a Cli configured with the given writer and version string.
func New(writer io.Writer, version string) *Cli {
	return &Cli{
		Writer: writer,
		Cmd: &cli.Command{
			Writer:                 writer,
			Version:                version,
			Name:                   appName,
			Usage:                  "a tool that converts HCL files to your favourite CICD provider.",
			UseShortOptionHandling: true,
			Copyright:              copyright,
			Authors: formattedAuthors([]mail.Address{
				{Name: "João Guimarães", Address: "joao.guimaraes@yld.com"},
			}),
		},
	}
}

func formattedAuthors(authors []mail.Address) []any {
	formatted := make([]any, 0, len(authors))

	for _, author := range authors {
		switch {
		case author.Name != "" && author.Address != "":
			formatted = append(formatted, fmt.Sprintf("\"%s\" <%s>", author.Name, author.Address))
		case author.Address != "":
			formatted = append(formatted, author.Address)
		default:
			formatted = append(formatted, author.Name)
		}
	}

	return formatted
}

func (cmd *Cli) addProvider(p provider.Provider) *cli.Command {
	providerCmd := &cli.Command{
		Name:  p.GetProviderName(),
		Usage: p.GetDescription(),
		Commands: []*cli.Command{
			{
				Name:  "parse",
				Usage: p.GetParseDescription(),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					opts, warnings, err := toProviderOpts(cmd, p.GetProviderName(), "parse")
					if err != nil {
						return err
					}

					for _, warning := range warnings {
						_, _ = fmt.Fprintf(cmd.Root().ErrWriter, "warning: %s\n", warning)
					}

					if err := p.Parse(opts); err != nil {
						return err
					}

					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Value:   "",
						Usage:   "Load HCL from a `FILE`",
					},
					&cli.StringFlag{
						Name:    "directory",
						Aliases: []string{"d"},
						Value:   "",
						Usage:   "Load HCL files from a `DIRECTORY`",
					},
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Value:   false,
						Usage:   "Reads the directory recursively",
					},
					&cli.StringFlag{
						Name:  "output-directory",
						Value: "",
						Usage: "Parsed files (YAML) are created in `DIRECTORY`",
					},
					&cli.BoolFlag{
						Name:  "yml",
						Value: false,
						Usage: "Generate .yml files instead of .yaml",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Value: false,
						Usage: "Output to stdout",
					},
				},
			},
			{
				Name:  "unparse",
				Usage: p.GetUnparseDescription(),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					opts, warnings, err := toProviderOpts(cmd, p.GetProviderName(), "unparse")
					if err != nil {
						return err
					}

					for _, warning := range warnings {
						_, _ = fmt.Fprintf(cmd.Root().ErrWriter, "warning: %s\n", warning)
					}

					if err := p.Unparse(opts); err != nil {
						return err
					}

					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Value:   "",
						Usage:   "Load YAML from a `FILE`",
					},
					&cli.StringFlag{
						Name:    "directory",
						Aliases: []string{"d"},
						Value:   "",
						Usage:   "Load YAML files from a `DIRECTORY`",
					},
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Value:   false,
						Usage:   "Reads the directory recursively",
					},
					&cli.StringFlag{
						Name:  "output-directory",
						Value: "",
						Usage: "Parsed files (HCL) are created in `DIRECTORY`",
					},
					&cli.BoolFlag{
						Name:  "dry-run",
						Value: false,
						Usage: "Output to stdout",
					},
				},
			},
		},
	}

	providerCmd.Commands = append(providerCmd.Commands, cmd.assistCommand(p))

	if p.GetProviderName() == "github" {
		providerCmd.Commands = append(providerCmd.Commands, cmd.pinCommand())
		providerCmd.Commands = append(providerCmd.Commands, cmd.upgradeCommand(p))
	}

	return providerCmd
}
