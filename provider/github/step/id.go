// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"fmt"
	"regexp"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseId(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

		if !re.MatchString(value.AsString()) {
			return fmt.Errorf("unsupported type, expected string that must start with a letter or _ and contain only alphanumeric characters, -, or _")
		}

		s.Id = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string that must start with a letter or _ and contain only alphanumeric characters, -, or _, found %s", value.Type().FriendlyName())
	}
}

func (config *StepConfig) parseId(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.Id, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}
