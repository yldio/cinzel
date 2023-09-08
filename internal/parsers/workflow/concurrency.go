package workflow

type ConcurrencyConfig struct {
	Group            string `hcl:"group,attr" yaml:"group"`
	CancelInProgress *bool  `hcl:"cancel_in_progress,attr" yaml:"cancel-in-progress,omitempty"`
}

func (config *ConcurrencyConfig) Parse() (ConcurrencyConfig, error) {
	return *config, nil
}
