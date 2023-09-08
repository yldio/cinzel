package step

type IfConfig string

func (config *IfConfig) Parse() (IfConfig, error) {
	return *config, nil
}
