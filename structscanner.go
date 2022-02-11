package structscanner

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// StructScanner represents a mapping from database columns to a struct type,
// and can be used to scan database rows to struct instances while ignoring NULL
// values. This makes it easier to implement outer joins where many fields may
// be NULL, but assumes that the zero value of the struct fields is sensible
// given a NULL.
//
// Only database fields with `db:` tags are mapped.
type StructScanner struct {
	prefix          string
	layout          *structLayout
	mappedFields    []*field
	mappedFieldPtrs []interface{}
}

func (s *StructScanner) columnWithoutPrefix(name string) string {
	if strings.HasPrefix(name, s.prefix+".") {
		return name[len(s.prefix)+1:]
	}
	return name
}

func (s *StructScanner) mapColumns(rows *sql.Rows) error {
	if s.mappedFields != nil {
		return nil
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	s.mappedFieldPtrs = make([]interface{}, len(columns))
	s.mappedFields = make([]*field, len(columns))

	for i := range columns {
		field := s.layout.fieldsByName[s.columnWithoutPrefix(columns[i])]
		if field == nil {
			panic(fmt.Sprintf("no destination field for '%s'", columns[i]))
		}

		s.mappedFieldPtrs[i] = reflect.New(reflect.PtrTo(field.Type)).Interface()
		s.mappedFields[i] = field
	}

	return nil
}

func (s *StructScanner) Scan(rows *sql.Rows, destPtr interface{}) error {
	err := s.mapColumns(rows)
	if err != nil {
		return err
	}

	err = rows.Scan(s.mappedFieldPtrs...)
	if err != nil {
		return err
	}

	s.setFields(destPtr)

	return nil
}

func (s *StructScanner) setFields(destPtr interface{}) {
	destValue := reflect.ValueOf(destPtr).Elem()

	for i := range s.mappedFields {
		destField := destValue.FieldByIndex(s.mappedFields[i].Indices)

		instanceValue := reflect.ValueOf(s.mappedFieldPtrs[i]).Elem().Elem()

		if !instanceValue.IsValid() {
			destField.Set(reflect.Zero(destField.Type()))

		} else if destField.Kind() == reflect.Ptr {
			value := reflect.New(s.mappedFields[i].Type)
			value.Elem().Set(instanceValue)
			destField.Set(value)

		} else {
			destField.Set(instanceValue)
		}
	}
}

func For(structPtr interface{}, prefix string) *StructScanner {
	structType := reflect.TypeOf(structPtr)

	cached, ok := cachedLayouts.Load(structType)
	if !ok {
		cached, _ = cachedLayouts.LoadOrStore(structType, newStructLayout(structType))
	}

	scanner := &StructScanner{
		prefix: prefix,
		layout: cached.(*structLayout),
	}
	return scanner
}
