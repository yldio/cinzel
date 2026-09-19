// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package gitlab

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/yldio/cinzel/internal/hclcomment"
	"github.com/zclconf/go-cty/cty"
)

func writeAttributeAny(body *hclwrite.Body, attr string, raw any) error {
	return writeCommentedAttribute(body, attr, raw, nodeComment{})
}

// writeNullAttribute writes the keyword as an explicit null, for the few
// GitLab reads as "clear what this would otherwise inherit".
//
// writeCommentedAttribute drops a null instead of writing one, so the callers
// that mean it say so here.
func writeNullAttribute(body *hclwrite.Body, attr string) error {
	body.SetAttributeValue(attr, cty.NullVal(cty.DynamicPseudoType))

	return nil
}

// writeCommentedAttribute writes the attribute with whatever comments its YAML
// key carried: the head run on its own lines above it, the inline one after
// the value. An empty comment writes nothing and leaves the attribute as it
// was.
func writeCommentedAttribute(body *hclwrite.Body, attr string, raw any, comment nodeComment) error {
	// A null is only a legal spelling on the handful of keywords GitLab reads
	// as clearing an inherited value, and parse drops one written anywhere
	// else. Writing it here meant unparse emitted "image = null" for its own
	// parse to delete: the HCL from the first pass and the HCL from the second
	// differed by a line that carried nothing. The callers that do mean a null
	// use writeNullAttribute.
	if raw == nil {
		return nil
	}

	ctyValue, err := anyToCty(raw)
	if err != nil {
		return err
	}

	hclcomment.WriteLeading(body, comment.head)

	if comment.line == "" {
		body.SetAttributeValue(attr, ctyValue)

		return nil
	}

	// SetAttributeValue writes the value and nothing after it, so the comment
	// has to ride along with the expression tokens to land on the same line.
	body.SetAttributeRaw(attr, hclcomment.Trailing(hclwrite.TokensForValue(ctyValue), comment.line))

	return nil
}

// writeReferenceAttribute writes a single reference, e.g. job.build, rather
// than a list of them.
func writeReferenceAttribute(body *hclwrite.Body, attr string, root string, ref string) {
	body.SetAttributeRaw(attr, hclwrite.Tokens{
		{Type: hclsyntax.TokenIdent, Bytes: []byte(fmt.Sprintf("%s.%s", root, ref))},
	})
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
