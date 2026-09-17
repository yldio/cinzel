// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclcomment"
	"github.com/zclconf/go-cty/cty"
)

// PreDecode populates the step fields from a cty object value (YAML-to-HCL path).
func (s *Step) PreDecode(val cty.Value) error {
	if val.IsNull() || !val.IsKnown() {
		return errors.New("is empty or not known")
	}

	if !val.Type().IsObjectType() {
		return errors.New("not a valid type")
	}

	mapping := val.AsValueMap()

	valId, ok := mapping["id"]

	if ok {
		if err := s.parseId(valId); err != nil {
			return err
		}
	} else {
		// No "id" in the source. Recorded here because this is the only point
		// that still sees the YAML as written: further down the step carries a
		// block label, and a label is indistinguishable from an id the author
		// wrote.
		s.IgnoreId = true
	}

	valIf, ok := mapping["if"]

	if ok {
		if err := s.parseIf(valIf); err != nil {
			return err
		}
	}

	valName, ok := mapping["name"]

	if ok {
		if err := s.parseName(valName); err != nil {
			return err
		}
	}

	valUses, ok := mapping["uses"]

	if ok {
		if err := s.parseUses(valUses); err != nil {
			return err
		}
	}

	valRun, ok := mapping["run"]

	if ok {
		if err := s.parseRun(valRun); err != nil {
			return err
		}
	}

	valWorkingDirectory, ok := mapping["working-directory"]

	if ok {
		if err := s.parseWorkingDirectory(valWorkingDirectory); err != nil {
			return err
		}
	}

	valShell, ok := mapping["shell"]

	if ok {
		if err := s.parseShell(valShell); err != nil {
			return err
		}
	}

	valWith, ok := mapping["with"]

	if ok {
		if err := s.parseWith(valWith); err != nil {
			return err
		}
	}

	valEnv, ok := mapping["env"]

	if ok {
		if err := s.parseEnv(valEnv); err != nil {
			return err
		}
	}

	valContinueOnError, ok := mapping["continue-on-error"]

	if ok {
		if err := s.parseContinueOnError(valContinueOnError); err != nil {
			return err
		}
	}

	valTimeoutMinutes, ok := mapping["timeout-minutes"]

	if ok {
		if err := s.parseTimeoutMinutes(valTimeoutMinutes); err != nil {
			return err
		}
	}

	return nil
}

// openAttr prepares stepBody for the attribute emitted as key: a blank line
// between it and whatever came before, then whatever comment was written above
// it. Returns the trailing comment, which the caller writes with the value.
func (s *Step) openAttr(stepBody *hclwrite.Body, key string) string {
	if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
		stepBody.AppendNewline()
	}

	comment := s.Comments.At(key)
	hclcomment.WriteLeading(stepBody, comment.Head)

	return comment.Line
}

// Decode writes the step as an HCL block into the given body.
func (s *Step) Decode(body *hclwrite.Body, attr string) error {
	if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
		body.AppendNewline()
	}

	hclcomment.WriteLeading(body, s.Comments.Head)

	stepBlock := body.AppendNewBlock(attr, []string{s.Identifier})
	stepBody := stepBlock.Body()

	// A step read from YAML with no "id" of its own must not gain one when it
	// is converted back, and parse defaults a step with no id to its block
	// label. "ignore_id" is what turns that default off, so it is written here
	// rather than an "id" the source never had.
	if s.IgnoreId {
		stepBody.SetAttributeValue("ignore_id", cty.True)
	} else if s.Id != cty.NilVal {
		line := s.openAttr(stepBody, "id")
		stepBody.SetAttributeRaw("id", hclcomment.Trailing(hclwrite.TokensForValue(s.Id), line))
	}

	if s.If != cty.NilVal {
		line := s.openAttr(stepBody, "if")
		stepBody.SetAttributeRaw("if", hclcomment.Trailing(hclwrite.TokensForValue(s.If), line))
	}

	if s.Name != cty.NilVal {
		line := s.openAttr(stepBody, "name")
		stepBody.SetAttributeRaw("name", hclcomment.Trailing(hclwrite.TokensForValue(s.Name), line))
	}

	if s.Uses != cty.NilVal {
		line := s.openAttr(stepBody, "uses")
		parts := strings.SplitN(s.Uses.AsString(), "@", 2)

		usesBlock := stepBody.AppendNewBlock("uses", nil)
		usesBody := usesBlock.Body()
		usesBody.SetAttributeValue("action", cty.StringVal(parts[0]))

		if len(parts) == 2 {
			// The comment rides the version, not the block: that is where the
			// pin tag names what the SHA above it stands for, and where parse
			// reads it back from.
			version := hclwrite.TokensForValue(cty.StringVal(parts[1]))
			usesBody.SetAttributeRaw("version", hclcomment.Trailing(version, line))
		}
	}

	if s.Run != cty.NilVal {
		line := s.openAttr(stepBody, "run")
		runStr := s.Run.AsString()

		if strings.Contains(runStr, "\n") {
			// A heredoc ends on its closing marker's own line, so a comment
			// after the tokens would land past it rather than beside the
			// attribute. It goes above instead, which is a place it can be.
			hclcomment.WriteLeading(stepBody, line)
			stepBody.SetAttributeRaw("run", setAsHeredoc(runStr))
		} else {
			stepBody.SetAttributeRaw("run", hclcomment.Trailing(hclwrite.TokensForValue(s.Run), line))
		}
	}

	if s.WorkingDirectory != cty.NilVal {
		line := s.openAttr(stepBody, "working-directory")
		stepBody.SetAttributeRaw("working_directory", hclcomment.Trailing(hclwrite.TokensForValue(s.WorkingDirectory), line))
	}

	if s.Shell != cty.NilVal {
		line := s.openAttr(stepBody, "shell")
		stepBody.SetAttributeRaw("shell", hclcomment.Trailing(hclwrite.TokensForValue(s.Shell), line))
	}

	if s.With != cty.NilVal {
		withMap := s.With.AsValueMap()
		keys := make([]string, 0, len(withMap))

		for k := range withMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		entries := s.Comments.Nest("with")

		for _, key := range keys {
			if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
				stepBody.AppendNewline()
			}

			comment := entries[key]

			withBlock := stepBody.AppendNewBlock("with", nil)
			withBody := withBlock.Body()
			withBody.SetAttributeValue("name", cty.StringVal(key))

			// The entry is one block here but one value in the YAML, so both
			// its comments go on the value: that is the half a reader sees,
			// and it is where parse looks for them coming back.
			hclcomment.WriteLeading(withBody, comment.Head)

			value := hclwrite.TokensForValue(withMap[key])
			withBody.SetAttributeRaw("value", hclcomment.Trailing(value, comment.Line))
		}
	}

	if s.Env != cty.NilVal {
		envMap := s.Env.AsValueMap()
		keys := make([]string, 0, len(envMap))

		for k := range envMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		entries := s.Comments.Nest("env")

		for _, name := range keys {
			if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
				stepBody.AppendNewline()
			}

			comment := entries[name]

			envBlock := stepBody.AppendNewBlock("env", nil)
			envBody := envBlock.Body()
			envBody.SetAttributeValue("name", cty.StringVal(name))

			// The entry is one block here but one value in the YAML, so both
			// its comments go on the value: that is the half a reader sees,
			// and it is where parse looks for them coming back.
			hclcomment.WriteLeading(envBody, comment.Head)

			value := hclwrite.TokensForValue(envMap[name])
			envBody.SetAttributeRaw("value", hclcomment.Trailing(value, comment.Line))
		}
	}

	if s.ContinueOnError != cty.NilVal {
		line := s.openAttr(stepBody, "continue-on-error")
		stepBody.SetAttributeRaw("continue_on_error", hclcomment.Trailing(hclwrite.TokensForValue(s.ContinueOnError), line))
	}

	if s.TimeoutMinutes != cty.NilVal {
		line := s.openAttr(stepBody, "timeout-minutes")
		stepBody.SetAttributeRaw("timeout_minutes", hclcomment.Trailing(hclwrite.TokensForValue(s.TimeoutMinutes), line))
	}

	return nil
}

func setAsHeredoc(content string) hclwrite.Tokens {
	content = strings.Trim(content, "\n")
	lines := strings.Split(content, "\n")

	// Heredoc tokens, with each line's newline inside the token rather than in
	// a separate TokenNewline. hclwrite.Format re-indents and re-spaces tokens
	// it reads as structure, which rewrites a shell script as if it were HCL:
	// "${{" becomes "${ {", ">>" becomes "> >", "if [ -n" becomes "if[-n".
	// Typed this way the body is opaque to it. The "<<-" form is required
	// because Format indents the closing marker.
	marker := freeHeredocMarker(lines)

	tokens := hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenOHeredoc,
			Bytes: []byte("<<-" + marker + "\n"),
		},
	}

	for _, line := range lines {
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenStringLit,
			Bytes: []byte(escapeTemplateMarkers(line) + "\n"),
		})
	}

	tokens = append(tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenCHeredoc,
		Bytes: []byte(marker),
	})

	return tokens
}

// freeHeredocMarker returns a marker no line of the body would be read as, so
// a script writing its own heredoc does not close ours early. The "<<-" form
// matches a closing marker after stripping the line's indentation, so the
// comparison is against the trimmed line. escapeTemplateMarkers leaves a bare
// marker alone, so the raw lines are what to check.
func freeHeredocMarker(lines []string) string {
	taken := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		taken[strings.TrimSpace(line)] = struct{}{}
	}

	marker := "EOF"

	for i := 1; ; i++ {
		if _, clash := taken[marker]; !clash {
			return marker
		}

		marker = "EOF_" + strconv.Itoa(i)
	}
}

// escapeTemplateMarkers doubles the "$" and "%" that open an HCL template
// sequence. Format reads a heredoc body byte by byte, so an unescaped "${"
// puts it into template mode and it re-spaces the rest as HCL: a GitHub
// expression "${{ x }}" comes back as "${ { x } }". Doubling matches how
// hclwrite writes ordinary string attributes, and parse turns "$${{" back
// into "${{".
func escapeTemplateMarkers(line string) string {
	line = strings.ReplaceAll(line, "${", "$${")

	return strings.ReplaceAll(line, "%{", "%%{")
}
