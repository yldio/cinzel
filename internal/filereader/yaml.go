// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

func (read *Reader[T]) FromYaml(path string, recursive bool) ([]T, error) {
	if err := read.readPath(path, recursive, []string{".yaml", ".yml"}); err != nil {
		return nil, err
	}

	list := *new([]T)

	for _, yamlFile := range read.files {
		yamlBytes, err := os.ReadFile(yamlFile)
		if err != nil {
			return nil, err
		}

		var item T = *new(T)

		if err := yaml.Unmarshal(yamlBytes, &item); err != nil {
			return nil, err
		}

		fileBase := filepath.Base(yamlFile)

		filename := strings.TrimSuffix(fileBase, filepath.Ext(fileBase))

		item.Update(filename)

		list = append(list, item)
	}

	return list, nil
}
