package step

type ShellConfig string

func (config *ShellConfig) Parse() (string, error) {
	return string(*config), nil
}
