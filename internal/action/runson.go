// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"github.com/zclconf/go-cty/cty"
)

type RunsOnConfig struct {
	On      *cty.Value `hcl:"on,attr"`
	OnGroup *string    `hcl:"on_group,attr"`
}

type RunsOnGroupConfig string

func (config *RunsOnConfig) Parse() (any, error) {
	if config == nil {
		return nil, nil
	}

	if config.On != nil {
		return ParseCtyValue(*config.On, []string{
			cty.String.FriendlyName(),
			cty.EmptyTuple.FriendlyName(),
		})
	} else if config.OnGroup != nil {
		return map[string]any{"group": *config.OnGroup}, nil
	}

	return nil, nil
}
