// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

import (
	"errors"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type RunsConfig struct {
	On      cty.Value `hcl:"on,attr"`
	OnGroup string    `hcl:"on_group,attr"`
}

type RunsOnGroupConfig string

type RunsOn any

func (config *RunsConfig) Parse() (RunsOn, error) {
	if config.On != cty.NilVal {
		switch config.On.Type().FriendlyName() {
		case cty.String.FriendlyName():
			var val string

			err := gocty.FromCtyValue(config.On, &val)
			if err != nil {
				return nil, err
			}

			return val, err
		case "tuple":
			var val []string

			for _, item := range config.On.AsValueSlice() {
				var itemVal string

				err := gocty.FromCtyValue(item, &itemVal)
				if err != nil {
					return nil, err
				}

				val = append(val, itemVal)
			}

			return val, nil
		default:
			return nil, errors.New("unknown runs-on type")
		}
	} else if config.OnGroup != "" {
		return map[string]any{"group": config.OnGroup}, nil
	}

	return nil, nil
}
