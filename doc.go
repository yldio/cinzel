// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0
// Command cinzel converts between HCL and CI/CD pipeline YAML in both
// directions: parse turns HCL definitions into YAML, unparse turns existing
// YAML back into HCL. Mappings are provider-specific, with GitHub Actions
// (workflows, composite actions and standalone steps) and GitLab CI/CD
// pipelines supported. It also generates HCL from a prompt through an LLM
// (assist) and pins GitHub action versions to commit SHAs (pin, upgrade).
package main
