// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package filereader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// FromYaml reads YAML files from path and unmarshals each into a value of type T.
func (read *Reader[T]) FromYaml(path string, recursive bool) ([]T, error) {

	if err := read.readPath(path, recursive, []string{".yaml", ".yml"}); err != nil {

		return nil, err
	}

	var list []T

	for _, yamlFile := range read.files {
		yamlBytes, err := os.ReadFile(yamlFile)
		if err != nil {

			return nil, err
		}

		var item T

		if err := yaml.Unmarshal(yamlBytes, &item); err != nil {

			return nil, err
		}

		fileBase := filepath.Base(yamlFile)
		fileName := strings.TrimSuffix(fileBase, filepath.Ext(fileBase))

		item.Update(fileName)

		list = append(list, item)
	}

	return list, nil
}

// CtyYaml unmarshals YAML content via the cty type system into a value of type T.
func (read *Reader[T]) CtyYaml(content []byte) (T, error) {
	ctyType := cty.Map(cty.DynamicPseudoType)
	var item T

	val, err := ctyyaml.Unmarshal(content, ctyType)
	if err != nil {

		return item, err
	}

	if val.IsNull() || !val.IsKnown() {

		return item, nil
	}

	if !val.Type().IsMapType() {

		return item, fmt.Errorf("expected a map type, got: %s", val.Type().FriendlyName())
	}

	for k, v := range val.AsValueMap() {

		if v.IsNull() || !v.IsKnown() {

			return item, nil
		}

		if err := gocty.FromCtyValue(v, &item); err != nil {

			return item, err
		}

		item.Update(k)
	}

	return item, nil
}
