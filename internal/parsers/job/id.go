// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: AGPL-3.0-only

package job

type IdConfig struct {
	Id string `hcl:"id,label"`
}

func (config *IdConfig) Parse() (Job, error) {
	job := Job{
		Id: config.Id,
	}

	return job, nil
}
