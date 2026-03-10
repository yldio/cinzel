// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseWith(value cty.Value) error {
	if value.Type().IsObjectType() {
		s.With = value

		return nil
	}

	return fmt.Errorf("unsupported type, expected string that must start with a letter or _ and contain only alphanumeric characters, -, or _, found %s", value.Type().FriendlyName())
}

func (config *StepConfig) parseWith(hv *hclparser.HCLVars) (cty.Value, error) {
	if config.With == nil {
		return cty.NilVal, nil
	}

	return config.With.Parse(hv)
}
