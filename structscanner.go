package structscanner

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// StructScanner represents a mapping from database columns to a struct type,
// and can be used to scan database rows to struct instances.
//
// Struct fields with a `db:` tag are mapped using the name specified in the tag
// as the column name. Only fields with `db:` tags are mapped.
//
// NULL values from the database are mapped to zero values on the struct.
//
// A StructScanner is not safe for concurrent use by multiple Goroutines.
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

// Scan populates the specified struct from a database row. Fields in the struct
// are set according to values in the row, based on the column name and the
// `db:` tag of the field. NULL values in the row result in the corresponding
// fields being set to their zero value.
//
// Columns are mapped when a row is first scanned. It is not safe to call Scan
// on a StructScanner with a query returning different columns after the first
// call.
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
		instanceValue := reflect.ValueOf(s.mappedFieldPtrs[i]).Elem().Elem()
		s.setNestedField(destValue, s.mappedFields[i].Indices, instanceValue)
	}
}

func (s *StructScanner) setNestedField(root reflect.Value, pathIndices []int, value reflect.Value) {
	destField := root
	for i := range pathIndices {
		if destField.Kind() == reflect.Ptr {
			if destField.IsNil() {

				// We’re setting a NULL - don’t keep traversing
				if !value.IsValid() {
					return
				}

				// Instantiate a zero instance for the pointer
				newValue := reflect.New(destField.Type().Elem())
				destField.Set(newValue)
			}

			destField = destField.Elem()
		}

		destField = destField.Field(pathIndices[i])
	}

	if !value.IsValid() {
		destField.Set(reflect.Zero(destField.Type()))

	} else if destField.Kind() == reflect.Ptr {
		newValue := reflect.New(destField.Type().Elem())
		newValue.Elem().Set(value)
		destField.Set(newValue)

	} else {
		destField.Set(value)
	}
}

// For returns a StructScanner suitable for scanning a struct of the type
// given by structPtr.
//
// A prefix may be specified; struct fields are mapped assuming that the columns
// from the database have the specified prefix with a dot (.) separator.
func For(structPtr interface{}, prefix string) StructScanner {
	structType := reflect.TypeOf(structPtr)

	cached, ok := cachedLayouts.Load(structType)
	if !ok {
		cached, _ = cachedLayouts.LoadOrStore(structType, newStructLayout(structType))
	}

	scanner := StructScanner{
		prefix: prefix,
		layout: cached.(*structLayout),
	}
	return scanner
}
