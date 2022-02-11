package structscanner

import (
	"database/sql"
	"fmt"
	"reflect"
)

func Select(tx Queryer, dest interface{}, prefix string, query string, args ...interface{}) error {
	destType := reflect.TypeOf(dest)
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

		s := For(dest, prefix)
		return s.Scan(rows, dest)

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

		reflect.ValueOf(dest).Elem().Set(resultValue.Elem())

		return nil
	}

	return fmt.Errorf("destination must be pointer to struct or slice")
}
