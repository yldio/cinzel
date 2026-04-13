// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package action

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

// Uses holds the parsed action reference and optional version.
type Uses struct {
	Action  cty.Value `yaml:"action,omitempty"`
	Version cty.Value `yaml:"version,omitempty"`
}

// UsesConfig represents the HCL configuration for a uses block.
type UsesConfig struct {
	Action  hcl.Expression `hcl:"action,attr"`
	Version hcl.Expression `hcl:"version,attr"`
}

// UsesListConfig is a slice of UsesConfig decoded from HCL uses blocks.
type UsesListConfig []UsesConfig

func (s *Uses) parseAction(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		s.Action = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string, found %s", value.Type().FriendlyName())
	}
}

func (s *Uses) parseVersion(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		s.Version = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string, found %s", value.Type().FriendlyName())
	}
}

func (config *UsesConfig) parseAction(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.Action, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}

func (config *UsesConfig) parseVersion(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.Version, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}

// extractExprComment reads the source file at the position immediately after
// the expression's range end and returns any trailing # comment on the same
// line, or empty string if none is present.
func extractExprComment(expr hcl.Expression) string {
	if expr == nil {
		return ""
	}

	r := expr.Range()
	if r.Filename == "" {
		return ""
	}

	src, err := os.ReadFile(r.Filename)
	if err != nil || int(r.End.Byte) >= len(src) {
		return ""
	}

	rest := src[r.End.Byte:]
	newline := bytes.IndexByte(rest, '\n')

	if newline < 0 {
		newline = len(rest)
	}

	tail := strings.TrimSpace(string(rest[:newline]))

	if !strings.HasPrefix(tail, "#") {
		return ""
	}

	return tail
}

// Parse resolves the uses block into a single "action@version" cty string value
// and the inline comment on the version attribute (if any).
func (config *UsesListConfig) Parse(hv *hclparser.HCLVars) (cty.Value, string, error) {
	if config == nil {
		return cty.NilVal, "", nil
	}

	if len(*config) != 1 {
		return cty.NilVal, "", errors.New("should only exist one uses")
	}

	c := (*config)[0]

	parsedUses := Uses{}

	action, err := c.parseAction(hv)
	if err != nil {
		return cty.NilVal, "", err
	}

	if action != cty.NilVal {
		if err := parsedUses.parseAction(action); err != nil {
			return cty.NilVal, "", err
		}
	} else {
		return cty.NilVal, "", errors.New("action must be set")
	}

	version, err := c.parseVersion(hv)
	if err != nil {
		return cty.NilVal, "", err
	}

	comment := ""

	if version != cty.NilVal {
		if err := parsedUses.parseVersion(version); err != nil {
			return cty.NilVal, "", err
		}

		comment = extractExprComment(c.Version)
	}

	var uses string

	if parsedUses.Version != cty.NilVal {
		uses = fmt.Sprintf("%s@%s", parsedUses.Action.AsString(), parsedUses.Version.AsString())
	} else {
		uses = parsedUses.Action.AsString()
	}

	return cty.StringVal(uses), comment, nil
}
