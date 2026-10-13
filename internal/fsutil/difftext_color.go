// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// The two SGR sequences a diff needs, and the reset that closes them.
const (
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
	ansiReset = "\x1b[0m"
)

// UseColor reports whether w should be written to in colour.
//
// The writer itself is asked, not os.Stdout. cinzel writes through an
// injected io.Writer, so a run whose output is a buffer or a pipe gets plain
// text while the same code on a terminal gets colour, and a test does not
// have to care which stream the process happens to own.
//
// NO_COLOR wins over everything, set to anything at all, per no-color.org.
// TERM=dumb is honoured for the terminals that announce they cannot.
func UseColor(w io.Writer) bool {
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	file, ok := w.(*os.File)
	if !ok {
		return false
	}

	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}

// Colorize puts red on the removed lines of a diff and green on the added
// ones, leaving the context and the hunk headers alone.
//
// Only whole lines are coloured, and the reset closes each one, so a line
// broken across a narrow terminal does not bleed into the next. Returns the
// diff untouched when enabled is false.
func Colorize(diff string, enabled bool) string {
	if !enabled || diff == "" {
		return diff
	}

	lines := strings.Split(diff, "\n")

	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, markerRemoved):
			lines[i] = ansiRed + line + ansiReset
		case strings.HasPrefix(line, markerAdded):
			lines[i] = ansiGreen + line + ansiReset
		}
	}

	return strings.Join(lines, "\n")
}
