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
	"github.com/yldio/cinzel/internal/pin"
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

			aiName := c.String("ai")
			model := c.String("model")

			aiProvider, err := resolveAIProvider(aiName, "")
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

			noContext := c.Bool("no-context")
			contextDir := c.String("context-dir")

			if contextDir == "" {
				contextDir = "cinzel"
			}

			if !noContext {
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
				refineDir := c.String("from")
				if refineDir == "" {
					refineDir = latestAssistDir(outputDir)
				} else {
					if err := validateRelativePath(refineDir); err != nil {
						return fmt.Errorf("--from: %w", err)
					}

					refineDir = filepath.Join(outputDir, refineDir)
				}

				if refineDir == "" {
					return fmt.Errorf("nothing to refine — run assist --prompt first to generate output in %s", outputDir)
				}

				assistContext, _ := ai.StripHCLContext(refineDir)
				if assistContext == "" {
					return fmt.Errorf("nothing to refine in %s — no HCL files found", refineDir)
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

			dedupDir := contextDir
			if noContext {
				dedupDir = ""
			}

			sessionDir, err := cmd.unparseAndWrite(p, yamlContent, outputDir, dedupDir, dryRun)
			if err != nil {
				return err
			}

			if p.GetProviderName() == "github" && sessionDir != "" {
				_, _ = fmt.Fprintf(cmd.Writer, "Pinning action versions...\n")

				resolver := pin.NewCachedResolver(pin.NewGitHubResolver(""))

				results, pinErr := pin.PinDirectory(ctx, sessionDir, resolver, cmd.Writer, false)
				if pinErr != nil {
					_, _ = fmt.Fprintf(cmd.Writer, "warning: pin failed: %v\n", pinErr)
				} else {
					cmd.printPinSummary(results)
				}
			}

			return nil
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
				Name:  "from",
				Value: "",
				Usage: "Target a specific assist session folder for --refine (e.g. 20260317-150405)",
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
				Name:  "ai",
				Value: "anthropic",
				Usage: "AI provider: anthropic or openai",
			},
			&cli.StringFlag{
				Name:  "model",
				Value: "",
				Usage: "Model override (default: AI provider-specific)",
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

// unparseAndWrite returns the session directory path where output was written (empty if dry-run).
func (cmd *Cli) unparseAndWrite(p provider.Provider, yamlContent, outputDir, contextDir string, dryRun bool) (string, error) {
	tmpYAMLDir, err := os.MkdirTemp("", "cinzel-assist-yaml-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	defer os.RemoveAll(tmpYAMLDir)

	tmpHCLDir, err := os.MkdirTemp("", "cinzel-assist-hcl-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
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
			return "", fmt.Errorf("failed to write temp file: %w", err)
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

		return "", fmt.Errorf(
			"generated YAML could not be converted to HCL:\n%s\n\nRaw YAML (preview):\n%s\n\nTry refining your prompt",
			err, preview,
		)
	}

	merged, err := mergeHCLFiles(tmpHCLDir)
	if err != nil {
		return "", fmt.Errorf("failed to merge HCL files: %w", err)
	}

	if contextDir != "" {
		merged = deduplicateWithExisting(merged, contextDir)
	}

	if dryRun {
		_, _ = fmt.Fprintln(cmd.Writer, merged)

		return "", nil
	}

	timestamp := time.Now().Format("20060102-150405")
	sessionDir := filepath.Join(outputDir, timestamp)

	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	outPath := filepath.Join(sessionDir, "assist.hcl")

	if err := os.WriteFile(outPath, []byte(merged), 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	absPath, _ := filepath.Abs(sessionDir)
	_, _ = fmt.Fprintf(cmd.Writer, "HCL written to %s\n", absPath)

	return sessionDir, nil
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

// existingBlock maps a block's content to its source file.
type existingBlock struct {
	content  string
	filename string
}

// blockSignature extracts the type and labels from an HCL block string,
// e.g. `step "checkout" {` → `step "checkout"`.
func blockSignature(block string) string {
	line := strings.SplitN(block, "\n", 2)[0]
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, "{")

	return strings.TrimSpace(line)
}

// deduplicateWithExisting compares generated blocks against existing HCL files
// in contextDir. Identical blocks are replaced with a reference comment.
// Blocks with matching signatures but different content are kept with a note.
func deduplicateWithExisting(merged, contextDir string) string {
	entries, err := os.ReadDir(contextDir)
	if err != nil {
		return merged
	}

	// Build index of existing blocks: signature → existingBlock.
	existing := make(map[string]existingBlock)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(contextDir, entry.Name()))
		if err != nil {
			continue
		}

		for _, block := range splitHCLBlocksAST(content, entry.Name()) {
			block = strings.TrimSpace(block)
			if block == "" {
				continue
			}

			sig := blockSignature(block)
			if sig == "" {
				continue
			}

			existing[sig] = existingBlock{
				content:  block,
				filename: entry.Name(),
			}
		}
	}

	if len(existing) == 0 {
		return merged
	}

	// Compare each generated block against existing ones.
	generatedBlocks := splitHCLBlocksAST([]byte(merged), "assist.hcl")

	var result []string

	for _, block := range generatedBlocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		sig := blockSignature(block)

		eb, found := existing[sig]
		if !found {
			result = append(result, block)

			continue
		}

		if eb.content == block {
			// Identical — replace with reference comment.
			result = append(result, fmt.Sprintf("// reuses: %s from %s", sig, eb.filename))

			continue
		}

		// Same signature but different content — keep with note.
		result = append(result, fmt.Sprintf("// note: %s also exists in %s (different content)\n%s", sig, eb.filename, block))
	}

	return strings.Join(result, "\n\n") + "\n"
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

// latestAssistDir returns the path to the most recent timestamped subfolder
// in the given directory, or empty string if none exist.
func latestAssistDir(baseDir string) string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return ""
	}

	var latest string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Timestamped folders match YYYYMMDD-HHMMSS pattern.
		if len(name) != 15 {
			continue
		}

		if name > latest {
			latest = name
		}
	}

	if latest == "" {
		return ""
	}

	return filepath.Join(baseDir, latest)
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
