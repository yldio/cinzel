// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/ai"
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
	return &cli.Command{
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
			cmd.assistCommand(p),
		},
	}
}

const defaultAssistOutputDir = "cinzel/assist"

func (cmd *Cli) assistCommand(p provider.Provider) *cli.Command {
	return &cli.Command{
		Name:  "assist",
		Usage: "Generate HCL workflow definitions from a natural language prompt",
		Action: func(ctx context.Context, c *cli.Command) error {
			prompt := c.String("prompt")
			if prompt == "" {
				return fmt.Errorf("--prompt is required")
			}

			outputDir := c.String("output-directory")
			if outputDir == "" {
				outputDir = defaultAssistOutputDir
			}

			dryRun := c.Bool("dry-run")
			acknowledge := c.Bool("acknowledge")

			if !acknowledge {
				if err := confirmCost(cmd.Writer, os.Stdin); err != nil {
					return err
				}
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Generating workflow...\n")

			systemPrompt := ai.SystemPrompt(p.GetProviderName())

			response, err := ai.Generate(ctx, systemPrompt, prompt, "")
			if err != nil {
				return err
			}

			yamlContent := ai.StripFences(response)

			return cmd.unparseAndWrite(p, yamlContent, outputDir, dryRun)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "prompt",
				Aliases:  []string{"p"},
				Usage:    "Natural language description of the workflow",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output-directory",
				Value: "",
				Usage: "Generated HCL files are created in `DIRECTORY` (default: cinzel/assist)",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Value: false,
				Usage: "Output to stdout instead of writing files",
			},
			&cli.BoolFlag{
				Name:  "acknowledge",
				Value: false,
				Usage: "Bypass the cost confirmation prompt",
			},
		},
	}
}

func (cmd *Cli) unparseAndWrite(p provider.Provider, yamlContent, outputDir string, dryRun bool) error {
	docs := splitYAMLDocuments(yamlContent)

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		tmpFile, err := os.CreateTemp("", "cinzel-assist-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}

		tmpPath := tmpFile.Name()

		if _, err := tmpFile.WriteString(doc); err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)

			return fmt.Errorf("failed to write temp file: %w", err)
		}

		tmpFile.Close()

		err = p.Unparse(provider.ProviderOps{
			File:            tmpPath,
			OutputDirectory: outputDir,
			DryRun:          dryRun,
		})

		os.Remove(tmpPath)

		if err != nil {
			return fmt.Errorf(
				"generated YAML could not be converted to HCL (document %d):\n%s\n\nRaw YAML:\n%s\n\nTry refining your prompt",
				i+1, err, doc,
			)
		}
	}

	if !dryRun {
		absDir, _ := filepath.Abs(outputDir)
		_, _ = fmt.Fprintf(cmd.Writer, "HCL files written to %s\n", absDir)
	}

	return nil
}

func splitYAMLDocuments(s string) []string {
	var docs []string
	var current strings.Builder

	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == "---" && current.Len() > 0 {
			docs = append(docs, current.String())
			current.Reset()

			continue
		}

		current.WriteString(line)
		current.WriteString("\n")
	}

	if strings.TrimSpace(current.String()) != "" {
		docs = append(docs, current.String())
	}

	return docs
}

func confirmCost(w io.Writer, r io.Reader) error {
	_, _ = fmt.Fprintf(w, "This will call Anthropic (claude-sonnet-4-5-20250514). API usage will incur costs.\nContinue? [y/N] ")

	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer == "y" || answer == "yes" {
			return nil
		}
	}

	return fmt.Errorf("cancelled")
}
