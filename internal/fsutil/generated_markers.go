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
func PruneStaleGeneratedYAML(outputDir string, currentOutputs map[string]struct{}, provider string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	cleanOutputDir, err := filepath.Abs(outputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := filepath.Join(outputDir, entry.Name())
		cleanPath := filepath.Clean(filePath)

		if _, ok := currentOutputs[cleanPath]; ok {
			continue
		}

		isOwned, err := HasGeneratedMarker(cleanPath, provider)
		if err != nil {
			return err
		}

		if !isOwned {
			continue
		}

		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(absPath, cleanOutputDir+string(os.PathSeparator)) {
			continue
		}

		if err := os.Remove(absPath); err != nil {
			return err
		}
	}

	return nil
}
