// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"fmt"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseContinueOnError(value cty.Value) error {
	switch value.Type() {
	case cty.String:
		s.ContinueOnError = value

		return nil
	case cty.Bool:
		s.ContinueOnError = value

		return nil
	default:
		return fmt.Errorf("unsupported type, expected string or bool, found %s", value.Type().FriendlyName())
	}
}

func (config *StepConfig) parseContinueOnError(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.ContinueOnError, hv)

	if err := hp.Parse(); err != nil {

		return cty.NilVal, err
	}

	return hp.Result(), nil
}
