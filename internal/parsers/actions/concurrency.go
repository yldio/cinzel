package actions

type ConcurrencyConfig struct {
	Group            string `hcl:"group,attr"`
	CancelInProgress bool   `hcl:"cancel_in_progress,attr"`
}

func (concurrency *ConcurrencyConfig) ConvertFromHcl() (Concurrency, error) {
	if concurrency == nil {
		return Concurrency{}, nil
	}

	content := Concurrency{
		Group:            concurrency.Group,
		CancelInProgress: concurrency.CancelInProgress,
	}
	return content, nil
}

type Concurrency struct {
	Group            string
	CancelInProgress bool
}

type ConcurrencyYaml struct {
	Group            string `yaml:"group,omitempty"`
	CancelInProgress bool   `yaml:"cancel-in-progress,omitempty"`
}

func (concurrency *Concurrency) ConvertToYaml() (ConcurrencyYaml, error) {
	content := ConcurrencyYaml{
		Group:            concurrency.Group,
		CancelInProgress: concurrency.CancelInProgress,
	}

	return content, nil
}
