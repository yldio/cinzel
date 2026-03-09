// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseUses(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		s.Uses = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string that must start with a letter or _ and contain only alphanumeric characters, -, or _, found %s", value.Type().FriendlyName())
	}
}

func (config *StepConfig) parseUses(hv *hclparser.HCLVars) (cty.Value, error) {

	if config.Uses == nil {

		return cty.NilVal, nil
	}

	return config.Uses.Parse(hv)
}
