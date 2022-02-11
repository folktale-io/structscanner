package structscanner

import "reflect"

type field struct {
	Name    string
	Type    reflect.Type
	Indices []int
}
