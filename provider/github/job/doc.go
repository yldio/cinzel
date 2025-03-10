// Copyright 2026 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package job contains GitHub job domain models and validation logic.
//
// This package owns:
//   - parsed job modeling from provider parse flows,
//   - job-specific validation rules (uses/runs-on/steps/needs),
//   - strategy matrix normalization helpers.
//
// Emission and orchestration remain in provider/github root.
package job
