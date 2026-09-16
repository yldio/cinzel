// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/ai"
)

// The template holds no API key. A key belongs in the environment, where
// it is not written to disk and not picked up by a backup of this file.
// LoadConfig still reads an api_key someone put here by hand, and the
// environment wins over it.
const configTemplate = `# cinzel AI configuration
# API keys are read from the environment:
#   ANTHROPIC_API_KEY, OPENAI_API_KEY

ai:
  default: %s
  providers:
%s`

// providerDefaults renders the per-provider block of the config template from
// the models the code actually defaults to. The template used to restate them,
// so a new config started on whatever was current when the string was written.
func providerDefaults() string {
	models := ai.DefaultModels()
	names := make([]string, 0, len(models))

	for name := range models {
		names = append(names, name)
	}

	sort.Strings(names)

	var b strings.Builder

	for _, name := range names {
		fmt.Fprintf(&b, "    %s:\n      model: %s\n", name, models[name])
	}

	return b.String()
}

func (cmd *Cli) initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create or update the cinzel configuration file",
		Action: func(ctx context.Context, c *cli.Command) error {
			configDir, err := configPath()
			if err != nil {
				return err
			}

			configFile := filepath.Join(configDir, "config.yaml")

			// One scanner for the whole prompt sequence. A second one over the
			// same stdin starts with an empty buffer and drops whatever the
			// first had already read ahead, which silently left every answer
			// below blank when the overwrite prompt ran.
			scanner := bufio.NewScanner(os.Stdin)

			if _, err := os.Stat(configFile); err == nil {
				_, _ = fmt.Fprintf(cmd.Writer, "Config already exists at %s\n", configFile)
				_, _ = fmt.Fprintf(cmd.Writer, "Overwrite? [y/N] ")

				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer != "y" && answer != "yes" {
						return nil
					}
				}
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Default AI provider (anthropic/openai) [anthropic]: ")

			defaultProvider := "anthropic"

			if scanner.Scan() {
				if input := strings.TrimSpace(scanner.Text()); input != "" {
					defaultProvider = input
				}
			}

			if err := os.MkdirAll(configDir, 0700); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			content := fmt.Sprintf(configTemplate, defaultProvider, providerDefaults())

			if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			// WriteFile only applies the mode when it creates the file, so an
			// existing config kept whatever permissions it had while the line
			// below claimed 0600. A config written before this command stopped
			// asking for keys may still hold one.
			if err := os.Chmod(configFile, 0600); err != nil {
				return fmt.Errorf("failed to set config file permissions: %w", err)
			}

			absPath, _ := filepath.Abs(configFile)
			_, _ = fmt.Fprintf(cmd.Writer, "\nConfig written to %s (permissions: 0600)\n", absPath)
			_, _ = fmt.Fprintf(cmd.Writer, "Set your API key in the environment:\n")
			_, _ = fmt.Fprintf(cmd.Writer, "  export ANTHROPIC_API_KEY=sk-ant-...\n")
			_, _ = fmt.Fprintf(cmd.Writer, "  export OPENAI_API_KEY=sk-...\n")

			return nil
		},
	}
}

// configPath returns the OS-agnostic cinzel config directory.
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not determine config directory: %w", err)
	}

	return filepath.Join(dir, "cinzel"), nil
}
