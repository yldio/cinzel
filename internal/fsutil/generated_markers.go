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
	generatedSource    = "# cinzel-source: %s"
	sourcePrefix       = "# cinzel-source:"
	maxMarkerScanLines = 8
)

// PrependGeneratedMarker prepends standardized cinzel generation markers.
//
// source names the HCL file the content was generated from, so a later run can
// tell which of its own outputs it is responsible for. It is written relative
// to the working directory with forward slashes, by SourceKey. An empty source
// writes no such line, which is what every file generated before the line
// existed looks like: PruneStaleGeneratedYAML reads one of those the way it
// always did.
func PrependGeneratedMarker(content []byte, provider string, source string) []byte {
	providerHeader := fmt.Sprintf(generatedProvider, provider)
	prefix := generatedByHeader + "\n" + providerHeader + "\n"

	if source != "" {
		prefix += fmt.Sprintf(generatedSource, source) + "\n"
	}

	return append([]byte(prefix), content...)
}

// WithoutGeneratedMarker returns comment with the cinzel generation markers
// removed, and empty string if that is all it held.
//
// The markers are written at the top of every generated file, so a YAML reader
// hands them back as the comment above the file's first key. They are cinzel's
// own note and not something an author wrote, and carrying them into the HCL
// would copy them into the source a person edits, one more line on each
// roundtrip. The source line is one of them: it names the HCL file the YAML
// came from, so unparsing that YAML back would write the note into the very
// file it names.
func WithoutGeneratedMarker(comment string) string {
	kept := []string{}

	for _, line := range strings.Split(comment, "\n") {
		trimmed := strings.TrimSpace(line)

		if trimmed == generatedByHeader || strings.HasPrefix(trimmed, "# cinzel-provider:") || strings.HasPrefix(trimmed, sourcePrefix) {
			continue
		}

		kept = append(kept, line)
	}

	return strings.Join(kept, "\n")
}

// HasGeneratedMarker reports whether path has cinzel markers for provider.
//
// A file that cannot be read is not one of cinzel's. The markers go on the
// first two lines of everything cinzel writes, so a file it cannot get them
// from is someone else's, and the answer to "may this be deleted" is no. It
// used to be an error instead, and the error travelled up through the prune
// and ended the run: a file the author put in the output directory with a line
// over 64KB in it, which bufio.Scanner refuses, failed a parse that had already
// written its YAML, at exit 1, pointing the author at the issue tracker over
// their own file.
func HasGeneratedMarker(path, provider string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, nil
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

	// scanner.Err() is not consulted: a file that could not be read through is
	// one whose markers were not found, which is the same answer this returns.
	return false, nil
}

// SourceKey spells a path to an HCL file the one way both sides of the
// ownership question agree on: relative to the working directory, with forward
// slashes.
//
// The two sides arrive spelled differently. A run given an absolute directory
// gets absolute paths back from the parser, one given "./cinzel" gets
// "cinzel/a.hcl", and the key written into a generated file last week was
// produced from whichever the author used then. Comparing those as they come
// answers "not mine" for a file that is, and the prune then leaves a renamed
// workflow's output behind for good.
//
// A file outside the working directory is spelled absolutely instead. Relative
// to a directory it is not under, it comes back as a run of "..", and how many
// depends on how deep the caller happened to be standing: the same file read
// from two places would then be two different files to the prune.
//
// A path that will not relativize at all, on Windows where it names another
// volume, returns empty. Empty is the same answer as a file written before the
// line existed, so such a run prunes the way it always did rather than failing.
func SourceKey(path string) string {
	if path == "" {
		return ""
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return filepath.ToSlash(abs)
	}

	rel, err := filepath.Rel(cwd, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return filepath.ToSlash(abs)
	}

	return filepath.ToSlash(rel)
}

// ReadSourceSet is the set of files a run read, spelled the way SourceKey
// spells a recorded source so the two can be compared.
//
// Built from the sources map ParseHCLInput returns, which is keyed by the path
// each file was parsed from. That is already the read set; naming it here keeps
// the prune from having to know how the input was reached.
func ReadSourceSet(sources map[string][]byte) map[string]struct{} {
	set := make(map[string]struct{}, len(sources))

	for path := range sources {
		if key := SourceKey(path); key != "" {
			set[key] = struct{}{}
		}
	}

	return set
}

// GeneratedSource returns the HCL file recorded in path's markers, and whether
// one was recorded at all.
//
// The two are not the same answer. A file with no source line was written
// before the line existed, and there is nothing to judge it by; a file with an
// empty one would be a file cinzel could not name the source of. Only the first
// happens today, and PruneStaleGeneratedYAML treats it as the migration case.
//
// A file that cannot be read has no source, for the reason HasGeneratedMarker
// gives: an unreadable file is not one this tool may act on, and an error here
// would travel up and end a run that had already written its YAML.
func GeneratedSource(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCount++

		if after, found := strings.CutPrefix(line, sourcePrefix); found {
			return strings.TrimSpace(after), true
		}

		if lineCount >= maxMarkerScanLines {
			break
		}
	}

	return "", false
}

// sameAsCurrentOutput reports whether path is one of the files this run wrote,
// comparing what the filesystem calls the same file rather than how the name is
// spelled. Every current output was written moments ago, so one that cannot be
// stat'd is not the file being looked at.
func sameAsCurrentOutput(path string, currentOutputs map[string]struct{}) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	for candidate := range currentOutputs {
		other, err := os.Stat(candidate)
		if err != nil {
			continue
		}

		if os.SameFile(info, other) {
			return true
		}
	}

	return false
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
//
// readSources names the HCL files this run read, spelled by SourceKey. A
// marked file recording a source outside that set was generated from an HCL
// file this run never looked at, and is left alone: "github parse -f cd.hcl"
// used to delete the YAML ci.hcl had produced, because the marker said cinzel
// wrote it and nothing about who asked. A rename still cleans up after itself,
// since the renamed workflow and the output it replaces share the file they
// were declared in.
//
// A file recording no source at all predates the line and is judged the way it
// always was. A nil readSources means the caller tracks no sources and asks for
// that same behaviour throughout.
func PruneStaleGeneratedYAML(outputDir string, currentOutputs map[string]struct{}, provider string, readSources map[string]struct{}) error {
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

		// Asked before the path work below, because a file this run has no
		// claim on is not one to go on reasoning about.
		if readSources != nil {
			if source, recorded := GeneratedSource(cleanPath); recorded {
				if _, ours := readSources[source]; !ours {
					return nil
				}
			}
		}

		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(absPath, cleanOutputDir+string(os.PathSeparator)) {
			return nil
		}

		// Asked here, where the answer decides a delete, rather than on the
		// name above. A path differing only in case is one file on macOS and
		// Windows, and os.WriteFile keeps the name the directory already
		// holds: renaming a workflow from "Build" to "build" left the walk
		// reading "Build.yaml" while currentOutputs held "build.yaml", so the
		// file the run had just written was read as stale and removed. parse
		// exited 0 having produced nothing at all.
		if sameAsCurrentOutput(absPath, currentOutputs) {
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
