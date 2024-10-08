// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import "errors"

type WorkingDirectoryConfig string

func (config *WorkingDirectoryConfig) Parse() (string, error) {
	if config == nil {
		return "", nil
	}

	return string(*config), nil
}

func ParseWorkingDirectory(val any) (*string, error) {
	if val == nil {
		return nil, nil
	}

	var content string
	var ok bool

	switch value := val.(type) {
	case []any:
		content, ok = value[0].(string)
	case any:
		content, ok = value.(string)
	}

	if !ok {
		return nil, errors.New("could not parse working-directory")
	}

	return &content, nil
}
