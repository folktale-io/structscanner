package structscanner

import (
	"database/sql"
	"fmt"
	"reflect"
)

// Select performs a database query, then scans the results into destPtr, which
// must be a pointer to a struct or a slice.
//
// If the destination is a struct, a single row is scanned into the struct. If
// no rows are returned by the query, sql.ErrNoRows is returned.
//
// If the destination is a slice, all rows from the query are scanned and placed
// in the slice. If no rows are returned by the query, the destination will be
// an empty slice.
//
// Columns returned from the query are mapped to struct fields using their `db:`
// tags, and column names are assumed to begin with prefix when mapping.
func Select(tx Queryer, destPtr interface{}, prefix string, query string, args ...interface{}) error {
	destType := reflect.TypeOf(destPtr)
	if destType.Kind() != reflect.Ptr {
		panic("pointer destination expected")
	}
	destType = destType.Elem()

	rows, err := tx.Query(query, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	if destType.Kind() == reflect.Struct {
		if !rows.Next() {
			return sql.ErrNoRows
		}

		s := For(destPtr, prefix)
		return s.Scan(rows, destPtr)

	} else if destType.Kind() == reflect.Slice {
		elemType := destType.Elem()
		elemDest := reflect.New(elemType)
		resultValue := reflect.New(destType)

		s := For(elemDest.Interface(), prefix)

		for rows.Next() {
			err := s.Scan(rows, elemDest.Interface())
			if err != nil {
				return err
			}

			resultValue.Elem().Set(reflect.Append(resultValue.Elem(), elemDest.Elem()))
			elemDest = reflect.New(elemType)
		}

		reflect.ValueOf(destPtr).Elem().Set(resultValue.Elem())

		return nil
	}

	return fmt.Errorf("destination must be pointer to struct or slice")
}
