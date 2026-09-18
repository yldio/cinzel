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
	"unicode/utf8"

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
	truncatedSuffix        = "\n... (truncated)"
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

			cfg, configWarnings := ai.LoadConfig()
			for _, warning := range configWarnings {
				warnTo(cmd.Writer, warning)
			}

			aiName := cfg.ResolveProviderName(c.String("ai"))
			model := cfg.ResolveModel(aiName, c.String("model"))
			apiKey := cfg.ResolveAPIKey(aiName)

			aiProvider, err := resolveAIProvider(aiName, apiKey)
			if err != nil {
				return err
			}

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
					_, _ = fmt.Fprintf(cmd.Writer, "Including existing HCL structure as context. String values are replaced with \"...\", but file names, block types, block labels and attribute names are sent as written. Use --no-context to skip sending existing HCL to the AI provider.\n")
					systemPrompt += "\n\nExisting HCL structure (string values replaced with \"...\"):\n\n" + hclContext
				}

				if truncated {
					_, _ = fmt.Fprintf(cmd.Writer, "warning: HCL context truncated to fit token limit\n")
				}
			}

			if !acknowledge {
				if err := confirmCost(cmd.Writer, os.Stdin, aiProvider.Name(), model); err != nil {
					return err
				}
			}

			_, _ = fmt.Fprintf(cmd.Writer, "Generating workflow...\n")

			userPrompt := prompt

			if refine != "" {
				_, _ = fmt.Fprintf(cmd.Writer, "Including previous assist output as context. String values are replaced with \"...\", but file names, block types, block labels and attribute names are sent as written.\n")

				refinedSystem, refinedUser, err := buildRefinePrompt(refine, prompt, outputDir, c.String("from"))
				if err != nil {
					return err
				}

				systemPrompt += refinedSystem
				userPrompt = refinedUser
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
				Value: "",
				Usage: "AI provider: anthropic or openai (default: from config or anthropic)",
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

// buildRefinePrompt resolves the refine directory, loads previous output as
// context, and returns the additional system prompt and user prompt.
func buildRefinePrompt(refine, prompt, outputDir, from string) (string, string, error) {
	refineDir := from
	if refineDir == "" {
		refineDir = latestAssistDir(outputDir)
	} else {
		if err := validateRelativePath(refineDir); err != nil {
			return "", "", fmt.Errorf("--from: %w", err)
		}

		refineDir = filepath.Join(outputDir, refineDir)
	}

	if refineDir == "" {
		return "", "", fmt.Errorf("nothing to refine — run assist --prompt first to generate output in %s", outputDir)
	}

	assistContext, _ := ai.StripHCLContext(refineDir)
	if assistContext == "" {
		return "", "", fmt.Errorf("nothing to refine in %s — no HCL files found", refineDir)
	}

	systemAddition := "\n\nPrevious assist output (to be refined):\n\n" + assistContext

	userPrompt := refine
	if prompt != "" {
		userPrompt = refine + "\n\nOriginal request: " + prompt
	}

	return systemAddition, userPrompt, nil
}

// truncatePreview cuts s to at most maxLen bytes for display in an error.
//
// maxLen is a byte budget, so the cut lands wherever it falls, which on a
// multibyte rune is mid-rune: the partial encoding left behind is not a rune
// and reaches the terminal as U+FFFD. Drop it. An encoding is at most four
// bytes, so at most three trailing bytes can be a partial one.
func truncatePreview(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	cut := s[:maxLen]

	for range utf8.UTFMax - 1 {
		if r, size := utf8.DecodeLastRuneInString(cut); r != utf8.RuneError || size != 1 {
			break
		}

		cut = cut[:len(cut)-1]
	}

	return cut + truncatedSuffix
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
		preview := truncatePreview(yamlContent, maxRawYAMLErrorLen)

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
		var warnings []string

		merged, warnings = deduplicateWithExisting(merged, contextDir)

		for _, warning := range warnings {
			warnTo(cmd.Writer, warning)
		}
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
//
// Leading comments and blank lines are skipped. Taking the first line outright
// returned the comment on any block carrying one — every action pin writes an
// "// action tag" line — so the signature matched nothing and the block could
// never be recognised as one the context already holds.
func blockSignature(block string) string {
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		return strings.TrimSpace(strings.TrimSuffix(line, "{"))
	}

	return ""
}

// deduplicateWithExisting compares generated blocks against existing HCL files
// in contextDir. Identical blocks are replaced with a reference comment.
// Blocks with matching signatures but different content are kept with a note.
//
// Warnings name anything that could not be read. A directory nobody created is
// the normal case and is silent, but one that exists and cannot be read is a
// deduplication that did not happen: without a word here, assist repeats the
// blocks the user already has and nothing says why.
func deduplicateWithExisting(merged, contextDir string) (string, []string) {
	var warnings []string

	entries, err := os.ReadDir(contextDir)
	if err != nil {
		if !os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("%s could not be read (%v). Generated blocks were not compared against it.", contextDir, err))
		}

		return merged, warnings
	}

	// Build index of existing blocks: signature → existingBlock.
	existing := make(map[string]existingBlock)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".hcl") {
			continue
		}

		path := filepath.Join(contextDir, entry.Name())

		content, err := os.ReadFile(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s could not be read (%v). Generated blocks were not compared against it.", path, err))

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
		return merged, warnings
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

	return strings.Join(result, "\n\n") + "\n", warnings
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

// splitYAMLDocuments cuts a stream into its YAML documents.
//
// A separator has to start at column 0. Accepting an indented one cut a
// workflow in half whenever a "run: |" block held a line that read "---",
// which a heredoc or an embedded manifest does, and both halves then failed
// to convert.
func splitYAMLDocuments(s string) []string {
	var docs []string
	var current strings.Builder

	for _, line := range strings.Split(s, "\n") {
		if strings.TrimRight(line, " \t\r") == "---" && current.Len() > 0 {
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

		// Timestamped folders match YYYYMMDD-HHMMSS. Checking the length
		// alone took any 15-character directory, so an unrelated one sorting
		// above the real sessions was handed to --refine as the last one.
		if _, err := time.Parse("20060102-150405", name); err != nil {
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
//
// Only a leading element that is exactly ".." escapes anything. Testing the
// first two characters refused "..hidden" and "...x" as well, which are
// ordinary directory names.
func validateRelativePath(p string) error {
	if filepath.IsAbs(p) {
		return errAbsolutePath
	}

	cleaned := filepath.Clean(p)

	// Clean leaves any ".." it could not resolve at the front, so checking the
	// first element covers "a/../../b" too.
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return errPathTraversal
	}

	return nil
}
