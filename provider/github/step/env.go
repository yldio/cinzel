// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseEnv(value cty.Value) error {
	if value.Type().IsObjectType() {
		s.Env = value

		return nil
	}

	return fmt.Errorf("unsupported type, expected string that must start with a letter or _ and contain only alphanumeric characters, -, or _, found %s", value.Type().FriendlyName())
}

func (config *StepConfig) parseEnv(hv *hclparser.HCLVars) (cty.Value, error) {
	if config.Env == nil {
		return cty.NilVal, nil
	}

	return config.Env.Parse(hv)
}
