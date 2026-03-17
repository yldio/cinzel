// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/urfave/cli/v3"
	"github.com/yldio/cinzel/internal/ai"
	"github.com/yldio/cinzel/provider"
)

const (
	defaultAssistOutputDir = "cinzel/assist"
	maxRawYAMLErrorLen     = 500
)

func (cmd *Cli) assistCommand(p provider.Provider) *cli.Command {
	return &cli.Command{
		Name:  "assist",
		Usage: "Generate HCL workflow definitions from a natural language prompt",
		Action: func(ctx context.Context, c *cli.Command) error {
			prompt := c.String("prompt")
			refine := c.String("refine")

			if prompt == "" && refine == "" {
				return errPromptRequired
			}

			outputDir := c.String("output-directory")
			if outputDir == "" {
				outputDir = defaultAssistOutputDir
			}

			if err := validateRelativePath(outputDir); err != nil {
				return fmt.Errorf("--output-directory: %w", err)
			}

			dryRun := c.Bool("dry-run")
			acknowledge := c.Bool("acknowledge")

			aiProviderName := c.String("provider")
			model := c.String("model")

			aiProvider, err := resolveAIProvider(aiProviderName, "")
			if err != nil {
				return err
			}

			if !acknowledge {
				if err := confirmCost(cmd.Writer, os.Stdin, aiProvider.Name(), model); err != nil {
					return err
				}
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Generating workflow...\n")

			systemPrompt := ai.SystemPrompt(p.GetProviderName())

			if !c.Bool("no-context") {
				contextDir := c.String("context-dir")
				if contextDir == "" {
					contextDir = "cinzel"
				}

				if err := validateRelativePath(contextDir); err != nil {
					return fmt.Errorf("--context-dir: %w", err)
				}

				hclContext, truncated := ai.StripHCLContext(contextDir)
				if hclContext != "" {
					systemPrompt += "\n\nExisting HCL structure (values stripped for privacy):\n\n" + hclContext
				}

				if truncated {
					_, _ = fmt.Fprintf(cmd.Writer, "warning: HCL context truncated to fit token limit\n")
				}
			}

			userPrompt := prompt

			if refine != "" {
				assistContext, _ := ai.StripHCLContext(outputDir)
				if assistContext == "" {
					return fmt.Errorf("nothing to refine — run assist --prompt first to generate initial output in %s", outputDir)
				}

				systemPrompt += "\n\nPrevious assist output (to be refined):\n\n" + assistContext

				if prompt != "" {
					userPrompt = refine + "\n\nOriginal request: " + prompt
				} else {
					userPrompt = refine
				}
			}

			response, err := ai.GenerateWithTimeout(ctx, aiProvider, ai.GenerateRequest{
				SystemPrompt: systemPrompt,
				UserPrompt:   userPrompt,
				Model:        model,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Tokens used: %d (input: %d, output: %d)\n",
				response.TotalTokens(), response.InputTokens, response.OutputTokens)

			yamlContent := ai.StripFences(response.Text)

			return cmd.unparseAndWrite(p, yamlContent, outputDir, dryRun)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "prompt",
				Aliases: []string{"p"},
				Usage:   "Natural language description of the workflow",
			},
			&cli.StringFlag{
				Name:  "refine",
				Usage: "Refine previous assist output with additional instructions",
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
			&cli.StringFlag{
				Name:  "provider",
				Value: "anthropic",
				Usage: "AI provider: anthropic or openai",
			},
			&cli.StringFlag{
				Name:  "model",
				Value: "",
				Usage: "Model override (default: provider-specific)",
			},
			&cli.BoolFlag{
				Name:  "no-context",
				Value: false,
				Usage: "Skip injecting existing HCL as context",
			},
			&cli.StringFlag{
				Name:  "context-dir",
				Value: "",
				Usage: "Directory to read existing HCL from (default: cinzel)",
			},
		},
	}
}

func (cmd *Cli) unparseAndWrite(p provider.Provider, yamlContent, outputDir string, dryRun bool) error {
	tmpYAMLDir, err := os.MkdirTemp("", "cinzel-assist-yaml-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer os.RemoveAll(tmpYAMLDir)

	tmpHCLDir, err := os.MkdirTemp("", "cinzel-assist-hcl-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer os.RemoveAll(tmpHCLDir)

	docs := splitYAMLDocuments(yamlContent)

	for i, doc := range docs {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		tmpPath := filepath.Join(tmpYAMLDir, fmt.Sprintf("workflow-%d.yaml", i))

		if err := os.WriteFile(tmpPath, []byte(doc), 0600); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}
	}

	err = p.Unparse(provider.ProviderOps{
		Directory:       tmpYAMLDir,
		OutputDirectory: tmpHCLDir,
		DryRun:          false,
	})
	if err != nil {
		preview := yamlContent
		if len(preview) > maxRawYAMLErrorLen {
			preview = preview[:maxRawYAMLErrorLen] + "\n... (truncated)"
		}

		return fmt.Errorf(
			"generated YAML could not be converted to HCL:\n%s\n\nRaw YAML (preview):\n%s\n\nTry refining your prompt",
			err, preview,
		)
	}

	merged, err := mergeHCLFiles(tmpHCLDir)
	if err != nil {
		return fmt.Errorf("failed to merge HCL files: %w", err)
	}

	if dryRun {
		_, _ = fmt.Fprintln(cmd.Writer, merged)

		return nil
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	outPath := filepath.Join(outputDir, fmt.Sprintf("assist-%s.hcl", timestamp))

	if err := os.WriteFile(outPath, []byte(merged), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	absPath, _ := filepath.Abs(outPath)
	_, _ = fmt.Fprintf(cmd.Writer, "HCL written to %s\n", absPath)

	return nil
}

// mergeHCLFiles reads all HCL files in dir, parses them with the HCL AST,
// and returns a single merged output with duplicate blocks removed.
func mergeHCLFiles(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	seen := make(map[string]bool)

	var parts []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return "", err
		}

		for _, block := range splitHCLBlocksAST(content, entry.Name()) {
			block = strings.TrimSpace(block)
			if block == "" {
				continue
			}

			if seen[block] {
				continue
			}

			seen[block] = true
			parts = append(parts, block)
		}
	}

	return strings.Join(parts, "\n\n") + "\n", nil
}

// splitHCLBlocksAST uses the HCL write parser to split content into
// individual top-level blocks. This is robust against braces inside
// strings, comments, and heredocs.
func splitHCLBlocksAST(src []byte, filename string) []string {
	file, diags := hclwrite.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		// Fall back to raw content as a single block if parse fails.
		return []string{string(src)}
	}

	var blocks []string

	for _, block := range file.Body().Blocks() {
		blocks = append(blocks, strings.TrimSpace(string(block.BuildTokens(nil).Bytes())))
	}

	// Also capture top-level attributes (e.g., standalone assignments).
	attrs := file.Body().Attributes()

	attrNames := make([]string, 0, len(attrs))
	for name := range attrs {
		attrNames = append(attrNames, name)
	}

	sort.Strings(attrNames)

	for _, name := range attrNames {
		attr := attrs[name]
		blocks = append(blocks, strings.TrimSpace(string(attr.BuildTokens(nil).Bytes())))
	}

	return blocks
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

func resolveAIProvider(name, apiKey string) (ai.Provider, error) {
	switch strings.ToLower(name) {
	case "anthropic", "":
		return ai.NewAnthropic(apiKey)
	case "openai":
		return ai.NewOpenAI(apiKey)
	default:
		return nil, fmt.Errorf("unknown AI provider %q. Supported: anthropic, openai", name)
	}
}

func confirmCost(w io.Writer, r io.Reader, providerName, model string) error {
	if model == "" {
		model = "default"
	}

	_, _ = fmt.Fprintf(w, "This will call %s (%s). API usage will incur costs.\nContinue? [y/N] ", providerName, model)

	scanner := bufio.NewScanner(r)
	if scanner.Scan() {
		answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if answer == "y" || answer == "yes" {
			return nil
		}
	}

	return errCancelled
}

// validateRelativePath ensures a path is relative and does not escape the
// current working directory via ".." traversal or absolute paths.
func validateRelativePath(p string) error {
	if filepath.IsAbs(p) {
		return errAbsolutePath
	}

	cleaned := filepath.Clean(p)
	if strings.HasPrefix(cleaned, "..") {
		return errPathTraversal
	}

	return nil
}
