// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	generatedByHeader  = "# generated-by: cinzel"
	generatedProvider  = "# cinzel-provider: %s"
	maxMarkerScanLines = 8
)

// PrependGeneratedMarker prepends standardized cinzel generation markers.
func PrependGeneratedMarker(content []byte, provider string) []byte {
	providerHeader := fmt.Sprintf(generatedProvider, provider)
	prefix := generatedByHeader + "\n" + providerHeader + "\n"

	return append([]byte(prefix), content...)
}

// WithoutGeneratedMarker returns comment with the cinzel generation markers
// removed, and empty string if that is all it held.
//
// The markers are written at the top of every generated file, so a YAML reader
// hands them back as the comment above the file's first key. They are cinzel's
// own note and not something an author wrote, and carrying them into the HCL
// would copy them into the source a person edits, one more line on each
// roundtrip.
func WithoutGeneratedMarker(comment string) string {
	kept := []string{}

	for _, line := range strings.Split(comment, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == generatedByHeader || strings.HasPrefix(trimmed, "# cinzel-provider:") {
			continue
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}

// HasGeneratedMarker reports whether path has cinzel markers for provider.
func HasGeneratedMarker(path, provider string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	defer f.Close()

	providerHeader := fmt.Sprintf(generatedProvider, provider)
	scanner := bufio.NewScanner(f)
	foundGeneratedBy := false
	foundProvider := false
	lineCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		if line == generatedByHeader {
			foundGeneratedBy = true
		}

		if line == providerHeader {
			foundProvider = true
		}

		if foundGeneratedBy && foundProvider {
			return true, nil
		}

		if lineCount >= maxMarkerScanLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

// PruneStaleGeneratedYAML removes stale YAML files owned by provider.
//
// The whole tree under outputDir is walked, not only its top level. An output
// can sit in a subdirectory, either because a filename names one or because
// an action is written to its own folder, and a file left there was never
// reached: renaming an action kept the old one beside the new one for good.
//
// Only files carrying the provider's marker are removed, so anything the
// caller wrote by hand is left where it is. currentOutputs has to name every
// file this run produced, actions included, or a live file is read as stale
// and deleted.
func PruneStaleGeneratedYAML(outputDir string, currentOutputs map[string]struct{}, provider string) error {
	cleanOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}

	err = filepath.WalkDir(outputDir, func(current string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(current))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		cleanPath := filepath.Clean(current)

		if _, ok := currentOutputs[cleanPath]; ok {
			return nil
		}

		isOwned, err := HasGeneratedMarker(cleanPath, provider)
		if err != nil {
			return err
		}

		if !isOwned {
			return nil
		}

		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(absPath, cleanOutputDir+string(os.PathSeparator)) {
			return nil
		}

		return os.Remove(absPath)
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return nil
}
