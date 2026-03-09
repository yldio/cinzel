// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package yamlwriter

import (
	"strings"
)

// Updater is implemented by types that can provide a filename, validate, and post-process YAML output.
type Updater interface {
	GetFilename() string
	Validation() error
	PostChanges([]byte) []byte
}

// Writer marshals a collection of Updater values into named YAML files.
type Writer[T Updater] struct {
	content []T
}

// New returns a Writer that will process the given content items.
func New[T Updater](content []T) *Writer[T] {
	return &Writer[T]{
		content: content,
	}
}

// Do validates, marshals, and post-processes each item, returning a map of filename to YAML bytes.
func (config *Writer[T]) Do() (map[string][]byte, error) {
	yamls := make(map[string][]byte)

	for _, c := range config.content {
		if err := c.Validation(); err != nil {
			return map[string][]byte{}, err
		}

		out, err := Marshal(c)
		if err != nil {
			return map[string][]byte{}, err
		}

		filteredOut := c.PostChanges(out)

		yamls[strings.Join([]string{c.GetFilename(), "yaml"}, ".")] = filteredOut
	}

	return yamls, nil
}
