// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package variables

import (
	"fmt"
	"sync"
)

type Variables map[string]any

var lock = &sync.Mutex{}
var actoVariables *Variables

func Instance() *Variables {
	if actoVariables == nil {
		lock.Lock()
		defer lock.Unlock()

		if actoVariables == nil {
			actoVariables = &Variables{}
		}
	}

	return actoVariables
}

func (variables *Variables) Add(key string, value any) {
	(*variables)[key] = value
}

func (variables *Variables) GetValue(attr string, idx *int64) (any, error) {
	if idx == nil {
		return variables.GetValueByKey(attr)
	}

	return variables.GetValueByIndex(attr, *idx)
}

func (variables *Variables) GetValueByKey(key string) (any, error) {
	value, ok := (*variables)[key]
	if ok {
		return value, nil
	}

	return nil, fmt.Errorf("variable `%s` does not exist", key)
}

func (variables *Variables) GetValueByIndex(key string, idx int64) (any, error) {
	value, err := variables.GetValueByKey(key)
	if err != nil {
		return nil, err
	}

	return value, nil
}
