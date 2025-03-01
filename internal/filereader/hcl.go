// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/acto/internal/actoerrors"
)

func (read *Reader[T]) FromHCL(path string, recursive bool) (hcl.Body, error) {
	if err := read.readPath(path, recursive, []string{".hcl"}); err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()
	var bodies []hcl.Body

	for _, hclFile := range read.files {
		file, diags := parser.ParseHCLFile(hclFile)
		if diags.HasErrors() {
			return nil, actoerrors.ProcessHCLDiags(diags)
		}

		bodies = append(bodies, file.Body)
	}

	return hcl.MergeBodies(bodies), nil
}
