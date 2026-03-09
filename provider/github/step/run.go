// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseRun(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		s.Run = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string, found %s", value.Type().FriendlyName())
	}
}

func (config *StepConfig) parseRun(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.Run, hv)

	if err := hp.Parse(); err != nil {

		return cty.NilVal, err
	}

	return hp.Result(), nil
}
