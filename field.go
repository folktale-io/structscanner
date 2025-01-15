package structscanner

import "reflect"

var unknownField = &field{
	Type: reflect.TypeFor[any](),
}

type field struct {
	Name    string
	Type    reflect.Type
	Indices []int
}
