// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package workflow contains GitHub workflow domain models and validation logic.
//
// This package owns:
//   - parsed workflow modeling from provider parse flows,
//   - workflow YAML document modeling,
//   - workflow-level validation rules,
//   - trigger normalization/mapping helpers.
//
// Emission and orchestration remain in provider/github root.
package workflow
