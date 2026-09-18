// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/yldio/cinzel/internal/cinzelerror"
)

// RejectUnknown reports every argument and block in body that the decode did
// not take.
//
// A body declared with `hcl:",remain"` is carried for its source range, but it
// also collects whatever the struct above it did not declare. A misspelled
// attribute therefore went out silently dropped and the run still exited 0,
// while the same misspelling on a block with no such body was an error. Worse,
// a keyword spelled as an attribute where the schema wants a block — "uses" on
// a step — was dropped along with everything it meant, and the step went out
// empty.
func RejectUnknown(body hcl.Body) error {
	if body == nil {
		return nil
	}

	if _, diags := body.Content(&hcl.BodySchema{}); diags.HasErrors() {
		return cinzelerror.ProcessHCLDiags(diags)
	}

	return nil
}
