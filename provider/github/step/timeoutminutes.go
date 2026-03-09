// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package step

import (
	"fmt"
	"math/big"

	"github.com/yldio/cinzel/internal/hclparser"
	"github.com/zclconf/go-cty/cty"
)

func (s *Step) parseTimeoutMinutes(value cty.Value) error {
	switch value.Type() {
	case cty.Number:
		bf := value.AsBigFloat()

		if !bf.IsInt() {
			return fmt.Errorf("unsupported type, expected positive number")
		}

		intVal, acc := bf.Uint64()

		if acc != big.Exact {
			return fmt.Errorf("unsupported type, expected positive number")
		}

		s.TimeoutMinutes = cty.NumberUIntVal(intVal)

		return nil
	default:
		return fmt.Errorf("unsupported type, expected positive number but got %s", value.Type().FriendlyName())
	}
}

func (config *StepConfig) parseTimeoutMinutes(hv *hclparser.HCLVars) (cty.Value, error) {
	hp := hclparser.New(config.TimeoutMinutes, hv)

	if err := hp.Parse(); err != nil {
		return cty.NilVal, err
	}

	return hp.Result(), nil
}
