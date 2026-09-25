// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package unescape

import (
	"regexp"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Unicode replaces \uXXXX and \UXXXXXXXX escape sequences with their raw UTF-8
// equivalents for characters above U+009F. Both hclwrite and gopkg.in/yaml.v3
// escape characters that are printable in their output format: hclwrite escapes
// any rune Go's unicode.IsPrint rejects, which includes category-Cf characters
// such as U+200D (ZWJ), and yaml.v3's is_printable helper only recognises
// 3-byte UTF-8, so supplementary-plane characters (emoji) are escaped too.
//
// An escape the source itself contained is left alone. Two things say so. A
// writer emits a literal backslash doubled, so a run of backslashes of even
// length before "uXXXX" is text rather than an escape. And the rune has to be
// one a writer would have escaped: see escaped.
func Unicode(src []byte) []byte {
	return reEscape.ReplaceAllFunc(src, func(match []byte) []byte {
		// Count the backslashes the match ends up owning: with an even number,
		// the last one is itself escaped and "uXXXX" is ordinary text.
		slashes := len(match) - len(trimSlashes(match))
		if slashes%2 == 0 {
			return match
		}

		digits := match[slashes+1:]

		n, err := strconv.ParseInt(string(digits), 16, 32)
		if err != nil || n <= 0x9F || !utf8.ValidRune(rune(n)) || !escaped(rune(n)) {
			return match
		}

		var buf [utf8.UTFMax]byte
		l := utf8.EncodeRune(buf[:], rune(n))

		return append(append([]byte(nil), match[:slashes-1]...), buf[:l]...)
	})
}

func trimSlashes(b []byte) []byte {
	for i, c := range b {
		if c != '\\' {
			return b[i:]
		}
	}

	return nil
}

// The leading run of backslashes is part of the match so their parity can be
// checked: without it the regex would find the second backslash of a literal
// "\\uXXXX" and rewrite text the author wrote.
var reEscape = regexp.MustCompile(`\\+U[0-9A-Fa-f]{8}|\\+u[0-9A-Fa-f]{4}`)

// escaped reports whether a writer would have escaped r, which is the other
// half of telling an escape from the same six characters typed by hand.
//
// Parity alone is not enough, because only a double-quoted scalar doubles a
// literal backslash. A plain or single-quoted YAML scalar leaves it single, and
// an HCL heredoc does the same, so a shell line reading `printf \u00e9` arrived
// here at odd parity and had those six characters replaced by the one they
// name, at exit 0, on a value nothing had escaped. Quoting the value instead
// would not reach the heredoc, which has no escapes at all.
//
// What the two writers escape is narrow and known: hclwrite escapes a rune Go's
// unicode.IsPrint rejects, and yaml.v3 escapes those plus anything above the
// BMP, since its printable check only reads three-byte UTF-8. Nothing outside
// that ever arrives as a real escape, so anything outside it is text.
func escaped(r rune) bool {
	return !unicode.IsPrint(r) || r > 0xFFFF
}
