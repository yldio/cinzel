// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"
)

const configTemplate = `# cinzel AI configuration
# API keys are stored here (never commit this file)

ai:
  default: %s
  providers:
    anthropic:
      model: claude-sonnet-4-5-20250514
      api_key: "%s"
    openai:
      model: gpt-4o
      api_key: "%s"
`

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

			if _, err := os.Stat(configFile); err == nil {
				_, _ = fmt.Fprintf(cmd.Writer, "Config already exists at %s\n", configFile)
				_, _ = fmt.Fprintf(cmd.Writer, "Overwrite? [y/N] ")

				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
					if answer != "y" && answer != "yes" {
						return nil
					}
				}
			}

			scanner := bufio.NewScanner(os.Stdin)

			_, _ = fmt.Fprintf(cmd.Writer, "Default AI provider (anthropic/openai) [anthropic]: ")

			defaultProvider := "anthropic"

			if scanner.Scan() {
				if input := strings.TrimSpace(scanner.Text()); input != "" {
					defaultProvider = input
				}
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Anthropic API key (leave empty to skip): ")

			var anthropicKey string

			if scanner.Scan() {
				anthropicKey = strings.TrimSpace(scanner.Text())
			}

			_, _ = fmt.Fprintf(cmd.Writer, "OpenAI API key (leave empty to skip): ")

			var openaiKey string

			if scanner.Scan() {
				openaiKey = strings.TrimSpace(scanner.Text())
			}

			if err := os.MkdirAll(configDir, 0700); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			content := fmt.Sprintf(configTemplate, defaultProvider, anthropicKey, openaiKey)

			if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			absPath, _ := filepath.Abs(configFile)
			_, _ = fmt.Fprintf(cmd.Writer, "\nConfig written to %s (permissions: 0600)\n", absPath)
			_, _ = fmt.Fprintf(cmd.Writer, "You can also set keys via environment variables:\n")
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
