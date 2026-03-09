// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"errors"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
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

// Decode writes the step as an HCL block into the given body.
func (s *Step) Decode(body *hclwrite.Body, attr string) error {

	if len(body.Blocks()) > 0 || len(body.Attributes()) > 0 {
		body.AppendNewline()
	}

	stepBlock := body.AppendNewBlock(attr, []string{s.Identifier})
	stepBody := stepBlock.Body()

	if s.Id != cty.NilVal {
		stepBody.SetAttributeValue("id", s.Id)
	}

	if s.If != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue("if", s.If)
	}

	if s.Name != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue("name", s.Name)
	}

	if s.Uses != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		parts := strings.SplitN(s.Uses.AsString(), "@", 2)

		usesBlock := stepBody.AppendNewBlock("uses", nil)
		usesBody := usesBlock.Body()
		usesBody.SetAttributeValue("action", cty.StringVal(parts[0]))

		if len(parts) == 2 {
			usesBody.SetAttributeValue("version", cty.StringVal(parts[1]))
		}
	}

	if s.Run != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		runStr := s.Run.AsString()

		if strings.Contains(runStr, "\n") {
			tokens := setAsHeredoc(runStr)
			stepBody.SetAttributeRaw("run", tokens)
		} else {
			stepBody.SetAttributeValue("run", s.Run)
		}
	}

	if s.WorkingDirectory != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue("working_directory", s.WorkingDirectory)
	}

	if s.Shell != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue("shell", s.Shell)
	}

	if s.With != cty.NilVal {
		withMap := s.With.AsValueMap()
		keys := make([]string, 0, len(withMap))

		for k := range withMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, key := range keys {

			if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
				stepBody.AppendNewline()
			}

			withBlock := stepBody.AppendNewBlock("with", nil)
			withBody := withBlock.Body()
			withBody.SetAttributeValue("name", cty.StringVal(key))
			withBody.SetAttributeValue("value", withMap[key])
		}
	}

	if s.Env != cty.NilVal {
		envMap := s.Env.AsValueMap()
		keys := make([]string, 0, len(envMap))

		for k := range envMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, name := range keys {

			if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
				stepBody.AppendNewline()
			}

			envBlock := stepBody.AppendNewBlock("env", nil)
			envBody := envBlock.Body()
			envBody.SetAttributeValue("name", cty.StringVal(name))
			envBody.SetAttributeValue("value", envMap[name])
		}
	}

	if s.ContinueOnError != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue("continue_on_error", s.ContinueOnError)
	}

	if s.TimeoutMinutes != cty.NilVal {

		if len(stepBody.Blocks()) > 0 || len(stepBody.Attributes()) > 0 {
			stepBody.AppendNewline()
		}

		stepBody.SetAttributeValue("timeout_minutes", s.TimeoutMinutes)
	}

	return nil
}

func setAsHeredoc(content string) hclwrite.Tokens {
	content = strings.Trim(content, "\n")
	lines := strings.Split(content, "\n")

	tokens := hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenOQuote,
			Bytes: []byte("<<EOF"),
		},
		{
			Type:  hclsyntax.TokenNewline,
			Bytes: []byte("\n"),
		},
	}

	for _, line := range lines {
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenQuotedLit,
			Bytes: []byte(line),
		}, &hclwrite.Token{
			Type:         hclsyntax.TokenNewline,
			Bytes:        []byte("\n"),
			SpacesBefore: 0,
		})
	}

	tokens = append(tokens, &hclwrite.Token{
		Type:  hclsyntax.TokenCQuote,
		Bytes: []byte("EOF"),
	})

	return tokens
}
