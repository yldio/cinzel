package job

type TimeoutMinutesConfig uint16

func (config *TimeoutMinutesConfig) Parse() (uint16, error) {
	return uint16(*config), nil
}
