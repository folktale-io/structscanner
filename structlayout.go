package structscanner

import (
	"database/sql"
	"reflect"
	"sync"
	"time"
)

var cachedLayouts = &sync.Map{}

type structLayout struct {
	fields       []field
	fieldsByName map[string]*field
}

func findStructFields(t reflect.Type, parentPath string, parentFieldIndex []int, fields *[]field) {
	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
	}

	fieldCount := t.NumField()

	scannerInterface := reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	timeType := reflect.TypeOf((*time.Time)(nil)).Elem()

	for i := 0; i < fieldCount; i++ {
		f := t.Field(i)

		tag := f.Tag.Get("db")
		if tag == "" {
			continue
		}

		fieldPath := tag
		if parentPath != "" {
			fieldPath = parentPath + "." + fieldPath
		}

		fieldIndex := make([]int, len(parentFieldIndex)+1)
		copy(fieldIndex, parentFieldIndex)
		fieldIndex[len(fieldIndex)-1] = f.Index[0]

		fieldType := f.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct &&
			!reflect.PtrTo(fieldType).Implements(scannerInterface) &&
			fieldType != timeType {

			findStructFields(f.Type, fieldPath, fieldIndex, fields)

		} else {
			*fields = append(*fields, field{
				Name:    fieldPath,
				Type:    fieldType,
				Indices: fieldIndex,
			})
		}
	}
}

func newStructLayout(sType reflect.Type) *structLayout {
	sl := &structLayout{
		fieldsByName: make(map[string]*field),
	}

	findStructFields(sType, "", nil, &sl.fields)

	for i := range sl.fields {
		sl.fieldsByName[sl.fields[i].Name] = &sl.fields[i]
	}

	return sl
}
