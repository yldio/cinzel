// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseWorkingDirectory(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		s.WorkingDirectory = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string, found %s", value.Type().FriendlyName())
	}
}

func (config *StepConfig) parseWorkingDirectory(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.WorkingDirectory, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}
