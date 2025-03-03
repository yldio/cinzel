// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package yamlwriter

import (
	"strings"

	"github.com/goccy/go-yaml"
)

type Updater interface {
	Update(string)
	GetFilename() string
	Validation() error
	PostChanges([]byte) []byte
}

type Writer[T Updater] struct {
	content []T
}

func New[T Updater](content []T) *Writer[T] {
	return &Writer[T]{
		content: content,
	}
}

func (config *Writer[T]) Do() (map[string][]byte, error) {
	yamls := make(map[string][]byte)

	for _, c := range config.content {
		if err := c.Validation(); err != nil {
			return map[string][]byte{}, err
		}

		out, err := yaml.Marshal(c)
		if err != nil {
			return map[string][]byte{}, err
		}

		filteredOut := c.PostChanges(out)

		yamls[strings.Join([]string{c.GetFilename(), "yaml"}, ".")] = filteredOut
	}

	return yamls, nil
}
