// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package filereader

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/cinzel/internal/cinzelerror"
)

// FromHCL reads HCL files from path, parses them, and returns a merged HCL body.
func (read *Reader[T]) FromHCL(path string, recursive bool) (hcl.Body, error) {

	if err := read.readPath(path, recursive, []string{".hcl"}); err != nil {

		return nil, err
	}

	parser := hclparse.NewParser()
	var bodies []hcl.Body

	for _, hclFile := range read.files {
		file, diags := parser.ParseHCLFile(hclFile)

		if diags.HasErrors() {

			return nil, cinzelerror.ProcessHCLDiags(diags)
		}

		bodies = append(bodies, file.Body)
	}

	return hcl.MergeBodies(bodies), nil
}
