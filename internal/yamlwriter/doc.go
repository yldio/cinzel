// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0
// Package yamlwriter marshals Go structs into YAML, with special handling for
// cty.Value fields. It converts cty values through the go-cty-yaml serializer
// before producing the final YAML output.
package yamlwriter
