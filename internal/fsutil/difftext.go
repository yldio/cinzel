// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package fsutil

import (
	"fmt"
	"strings"
)

// diffContextLines is how many unchanged lines are shown either side of a
// change, the same three a unified diff conventionally carries.
const diffContextLines = 3

// NormalizeLineEndings rewrites CRLF as LF.
//
// Every generated file is written with LF, but the copy on disk came through
// a checkout, and git on Windows converts line endings by default. Comparing
// the two without this reports every line of every file as changed, which on
// a Windows runner makes the check useless rather than strict.
//
// A difference the author would recognise still fails: this only takes the
// line ending out of the comparison.
func NormalizeLineEndings(content []byte) []byte {
	return []byte(strings.ReplaceAll(string(content), "\r\n", "\n"))
}

// SameContent reports whether two files say the same thing, line endings
// aside.
func SameContent(a []byte, b []byte) bool {
	return string(NormalizeLineEndings(a)) == string(NormalizeLineEndings(b))
}

// UnifiedDiff renders the changes from want to got as a unified diff, with
// the usual "-" for what is there and "+" for what would be written.
//
// The line pairing is a longest common subsequence rather than a walk down
// both sides at once. Comparing line i against line i reports everything
// after an inserted line as changed, so a one-line addition near the top of a
// workflow prints the rest of the file and buries the line that moved. A CI
// log is read by someone looking for what to fix.
//
// The cost is a table the size of one file against the other, which these
// are small enough for: a generated workflow runs to a few hundred lines.
func UnifiedDiff(want []byte, got []byte) string {
	a := strings.Split(strings.TrimSuffix(string(NormalizeLineEndings(want)), "\n"), "\n")
	b := strings.Split(strings.TrimSuffix(string(NormalizeLineEndings(got)), "\n"), "\n")

	return renderHunks(a, b, commonSubsequence(a, b))
}

// commonSubsequence pairs the lines the two sides share, as indices into each.
func commonSubsequence(a []string, b []string) [][2]int {
	lengths := make([][]int, len(a)+1)

	for i := range lengths {
		lengths[i] = make([]int, len(b)+1)
	}

	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lengths[i][j] = lengths[i+1][j+1] + 1

				continue
			}

			if lengths[i+1][j] >= lengths[i][j+1] {
				lengths[i][j] = lengths[i+1][j]

				continue
			}

			lengths[i][j] = lengths[i][j+1]
		}
	}

	pairs := make([][2]int, 0, min(len(a), len(b)))

	for i, j := 0, 0; i < len(a) && j < len(b); {
		switch {
		case a[i] == b[j]:
			pairs = append(pairs, [2]int{i, j})
			i++
			j++
		case lengths[i+1][j] >= lengths[i][j+1]:
			i++
		default:
			j++
		}
	}

	return pairs
}

// diffLine is one rendered line and where it sits on each side.
type diffLine struct {
	text string
	a    int
	b    int
}

// renderHunks turns the pairing into unified diff text, showing only the
// lines that changed and the context around them.
func renderHunks(a []string, b []string, pairs [][2]int) string {
	lines := alignedLines(a, b, pairs)

	var out strings.Builder

	for start := 0; start < len(lines); {
		if lines[start].text[0] == ' ' {
			start++

			continue
		}

		from, to := hunkBounds(lines, start)

		writeHunkHeader(&out, lines[from:to])

		for _, line := range lines[from:to] {
			out.WriteString(line.text)
			out.WriteString("\n")
		}

		start = to
	}

	return out.String()
}

// alignedLines walks both sides in order, emitting a removal, an addition or
// a context line for each.
func alignedLines(a []string, b []string, pairs [][2]int) []diffLine {
	lines := make([]diffLine, 0, len(a)+len(b))
	i, j := 0, 0

	emitUpTo := func(nextA int, nextB int) {
		for i < nextA {
			lines = append(lines, diffLine{text: "-" + a[i], a: i, b: j})
			i++
		}

		for j < nextB {
			lines = append(lines, diffLine{text: "+" + b[j], a: i, b: j})
			j++
		}
	}

	for _, pair := range pairs {
		emitUpTo(pair[0], pair[1])

		lines = append(lines, diffLine{text: " " + a[i], a: i, b: j})
		i++
		j++
	}

	emitUpTo(len(a), len(b))

	return lines
}

// hunkBounds grows a change at start outwards by the context lines either
// side, stopping where the next change is too far away to join this one.
func hunkBounds(lines []diffLine, start int) (int, int) {
	from := max(start-diffContextLines, 0)
	to := start

	for to < len(lines) {
		if lines[to].text[0] != ' ' {
			to++

			continue
		}

		// A run of context long enough to close the hunk, unless another
		// change follows soon enough to be worth keeping in the same one.
		run := to

		for run < len(lines) && lines[run].text[0] == ' ' {
			run++
		}

		if run-to > diffContextLines*2 || run == len(lines) {
			to = min(to+diffContextLines, len(lines))

			break
		}

		to = run
	}

	return from, to
}

// writeHunkHeader writes the @@ line naming where the hunk sits in each file.
func writeHunkHeader(out *strings.Builder, hunk []diffLine) {
	countA, countB := 0, 0

	for _, line := range hunk {
		switch line.text[0] {
		case '-':
			countA++
		case '+':
			countB++
		default:
			countA++
			countB++
		}
	}

	fmt.Fprintf(out, "@@ -%d,%d +%d,%d @@\n", hunk[0].a+1, countA, hunk[0].b+1, countB)
}
