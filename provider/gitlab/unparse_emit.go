// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

package gitlab

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func writeAttributeAny(body *hclwrite.Body, attr string, raw any) error {
	ctyValue, err := anyToCty(raw)
	if err != nil {
		return err
	}

	body.SetAttributeValue(attr, ctyValue)

	return nil
}

func writeReferenceListAttribute(body *hclwrite.Body, attr string, root string, refs []string) error {
	if len(refs) == 0 {
		return nil
	}

	tokens := hclwrite.Tokens{{Type: hclsyntax.TokenOBrack, Bytes: []byte("[")}, {Type: hclsyntax.TokenNewline, Bytes: []byte("\n")}}

	for _, ref := range refs {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("%s.%s", root, ref))})
		tokens = append(tokens,
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		)
	}

	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte("]")})

	body.SetAttributeRaw(attr, tokens)

	return nil
}

func writeScopedReferenceListAttribute(body *hclwrite.Body, attr string, roots []string, refs []string) error {
	if len(roots) != len(refs) {
		return fmt.Errorf("reference roots and refs length mismatch")
	}

	if len(refs) == 0 {
		return nil
	}

	tokens := hclwrite.Tokens{{Type: hclsyntax.TokenOBrack, Bytes: []byte("[")}, {Type: hclsyntax.TokenNewline, Bytes: []byte("\n")}}

	for i, ref := range refs {
		tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("%s.%s", roots[i], ref))})
		tokens = append(tokens,
			&hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(",")},
			&hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
		)
	}

	tokens = append(tokens, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte("]")})

	body.SetAttributeRaw(attr, tokens)

	return nil
}
