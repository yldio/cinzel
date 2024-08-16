// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package action

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

func Parse(config hcl.Expression, hclType string) ([]string, error) {
	if config == nil {
		return nil, nil
	}

	val, diags := config.Value(nil)
	if diags.HasErrors() {
		// return nil, errors.New(diags[0].Detail)
	}
	if val.IsNull() {
		return nil, nil
	}

	exprs, diags := hcl.ExprList(config)
	if diags.HasErrors() {
		return nil, errors.New(diags[0].Detail)
	}

	ids := []string{}

	for _, expr := range exprs {
		traversal, diags := hcl.AbsTraversalForExpr(expr)
		if diags.HasErrors() {
			return nil, errors.New(diags[0].Detail)
		}

		for _, traverser := range traversal {
			switch tJob := traverser.(type) {
			case hcl.TraverseRoot:
				if tJob.Name != hclType {
					return nil, fmt.Errorf("%ss require a %s relationship only", hclType, hclType)
				}
			case hcl.TraverseAttr:
				ids = append(ids, tJob.Name)
			}
		}
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("requires at least one %s", hclType)
	}

	return nil, nil
}
