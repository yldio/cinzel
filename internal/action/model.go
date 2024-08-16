package action

type A struct {
	Name string
}

func (config *A) Parse() {}

type B string

func (config *B) Parse() {}

type C map[string]bool

func (config *C) Parse() {}

type ActionMapper interface {
	A | B | C
	Parse()
}

func Parser[K comparable, Y ActionMapper](any) K {
	var yml Y
	var kapa K

	yml.Parse()

	return kapa
}
