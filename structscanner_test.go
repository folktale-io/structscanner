package structscanner

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStructScanner(t *testing.T) {

	db, dbMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error setting up mock DB: %v", err)
	}
	defer db.Close()

	type NestedStruct struct {
		NestedStringValue    string  `db:"nested_string_value"`
		NestedStringPtrValue *string `db:"nested_string_ptr_value"`
	}

	type TestStruct struct {
		StringValue    string        `db:"string_value"`
		StringPtrValue *string       `db:"string_ptr_value"`
		IntValue       int           `db:"int_value"`
		IntPtrValue    *int          `db:"int_ptr_value"`
		BoolValue      bool          `db:"bool_value"`
		BoolPtrValue   *bool         `db:"bool_ptr_value"`
		TimeValue      time.Time     `db:"time_value"`
		TimePtrValue   *time.Time    `db:"time_ptr_value"`
		NullTimeValue  sql.NullTime  `db:"null_time_value"`
		StructValue    NestedStruct  `db:"struct_value"`
		StructPtr      *NestedStruct `db:"struct_ptr"`
	}

	t.Run("Scan", func(t *testing.T) {

		t.Run("maps fields from query to struct", func(t *testing.T) {
			ss := For((*TestStruct)(nil), "prefix")

			mockQuery := fmt.Sprintf("some query")

			dbMock.ExpectQuery(mockQuery).WillReturnRows(
				sqlmock.NewRows([]string{
					"prefix.string_value",
					"prefix.string_ptr_value",
					"prefix.int_value",
					"prefix.int_ptr_value",
					"prefix.bool_value",
					"prefix.bool_ptr_value",
					"prefix.time_value",
					"prefix.time_ptr_value",
					"prefix.null_time_value",
					"prefix.struct_value.nested_string_value",
					"prefix.struct_value.nested_string_ptr_value",
					"prefix.struct_ptr.nested_string_value",
					"prefix.struct_ptr.nested_string_ptr_value",
				}).AddRow(
					"string value",
					"string ptr value",
					123,
					123,
					true,
					true,
					time.Date(2022, 02, 11, 12, 13, 14, 0, time.UTC),
					time.Date(2022, 02, 11, 12, 13, 14, 0, time.UTC),
					time.Date(2022, 02, 11, 12, 13, 14, 0, time.UTC),
					"nested string value 1",
					"nested string ptr value 1",
					"nested string value 2",
					"nested string ptr value 2",
				),
			)

			rows, err := db.Query(mockQuery)
			if err != nil {
				t.Fatalf("Error executing query: %v", err)
			}

			if !rows.Next() {
				t.Fatalf("Expected one row but got none")
			}

			var result TestStruct

			err = ss.Scan(rows, &result)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := "string value", result.StringValue; expected != actual {
				t.Errorf("Expected string value '%s' but got '%s'", expected, actual)
			}
			if expected, actual := "string ptr value", result.StringPtrValue; expected != *actual {
				t.Errorf("Expected string pointer value '%s' but got '%s'", expected, *actual)
			}
			if expected, actual := 123, result.IntValue; expected != actual {
				t.Errorf("Expected int value %d but got %d", expected, actual)
			}
			if expected, actual := 123, result.IntPtrValue; expected != *actual {
				t.Errorf("Expected int pointer value %d but got %d", expected, *actual)
			}
			if actual := result.BoolValue; !actual {
				t.Errorf("Expected bool value to be true but wasn’t")
			}
			if actual := result.BoolPtrValue; !(*actual) {
				t.Errorf("Expected bool pointer value to be true but wasn’t")
			}
			if expected, actual := time.Date(2022, 2, 11, 12, 13, 14, 0, time.UTC), result.TimeValue; expected != actual {
				t.Errorf("Expected time value %v but got %v", expected, actual)
			}
			if expected, actual := time.Date(2022, 2, 11, 12, 13, 14, 0, time.UTC), result.TimePtrValue; expected != *actual {
				t.Errorf("Expected time pointer value %v but got %v", expected, *actual)
			}
			if expected, actual := time.Date(2022, 2, 11, 12, 13, 14, 0, time.UTC), result.NullTimeValue.Time; expected != actual {
				t.Errorf("Expected nulltime value %v but got %v", expected, actual)
			}
			if expected, actual := "nested string value 1", result.StructValue.NestedStringValue; expected != actual {
				t.Errorf("Expected nested string value '%s' but got '%s'", expected, actual)
			}
			if expected, actual := "nested string ptr value 1", result.StructValue.NestedStringPtrValue; actual == nil || expected != *actual {
				t.Errorf("Expected nested string pointer value '%s' but got %v", expected, actual)
			}

			if actual := result.StructPtr; actual == nil {
				t.Errorf("Expected nested struct pointer to have been initialised but wasn’t")
			} else {
				if expected, actual := "nested string value 2", result.StructPtr.NestedStringValue; expected != actual {
					t.Errorf("Expected struct pointer to have nested string value '%s' but got '%s'", expected, actual)
				}
				if expected, actual := "nested string ptr value 2", result.StructPtr.NestedStringPtrValue; actual == nil || expected != *actual {
					t.Errorf("Expected struct pointer to have nested string ptr value '%s' but got %v", expected, actual)
				}
			}
		})

		t.Run("instantiates pointer fields when a nested field is set to non-NULL", func(t *testing.T) {
			ss := For((*TestStruct)(nil), "prefix")

			mockQuery := fmt.Sprintf("some query")

			dbMock.ExpectQuery(mockQuery).WillReturnRows(
				sqlmock.NewRows([]string{
					"prefix.struct_ptr.nested_string_ptr_value",
				}).AddRow(
					"nested string ptr value 2",
				),
			)

			rows, err := db.Query(mockQuery)
			if err != nil {
				t.Fatalf("Error executing query: %v", err)
			}

			if !rows.Next() {
				t.Fatalf("Expected one row but got none")
			}

			var result TestStruct

			err = ss.Scan(rows, &result)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if actual := result.StructPtr; actual == nil {
				t.Errorf("Expected nested struct pointer to have been initialised but wasn’t")
			} else {
				if expected, actual := "", result.StructPtr.NestedStringValue; expected != actual {
					t.Errorf("Expected struct pointer to have nested string value '%s' but got '%s'", expected, actual)
				}
				if expected, actual := "nested string ptr value 2", result.StructPtr.NestedStringPtrValue; actual == nil || expected != *actual {
					t.Errorf("Expected struct pointer to have nested string ptr value '%s' but got %v", expected, actual)
				}
			}
		})

		t.Run("maps NULLs to zero values in struct", func(t *testing.T) {
			ss := For((*TestStruct)(nil), "")

			stringValue := "string value"
			intValue := 123
			boolValue := true
			timeValue := time.Now().UTC()
			nestedStringValue := "nested string value"

			result := TestStruct{
				StringValue:    stringValue,
				StringPtrValue: &stringValue,
				IntValue:       intValue,
				IntPtrValue:    &intValue,
				BoolValue:      true,
				BoolPtrValue:   &boolValue,
				TimeValue:      timeValue,
				TimePtrValue:   &timeValue,
				NullTimeValue:  sql.NullTime{Time: timeValue, Valid: true},
			}
			result.StructValue.NestedStringValue = nestedStringValue
			result.StructValue.NestedStringPtrValue = &nestedStringValue

			mockQuery := fmt.Sprintf("some query")

			dbMock.ExpectQuery(mockQuery).WillReturnRows(
				sqlmock.NewRows([]string{
					"string_value",
					"string_ptr_value",
					"int_value",
					"int_ptr_value",
					"bool_value",
					"bool_ptr_value",
					"time_value",
					"time_ptr_value",
					"null_time_value",
					"struct_value.nested_string_value",
					"struct_value.nested_string_ptr_value",
					"struct_ptr.nested_string_ptr_value",
				}).AddRow(
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			)

			rows, err := db.Query(mockQuery)
			if err != nil {
				t.Fatalf("Error executing query: %v", err)
			}

			if !rows.Next() {
				t.Fatalf("Expected one row but got none")
			}

			err = ss.Scan(rows, &result)
			if err != nil {
				t.Fatalf("Expected success but got error: %v", err)
			}

			if expected, actual := "", result.StringValue; expected != actual {
				t.Errorf("Expected string value '%s' but got '%s'", expected, actual)
			}
			if expected, actual := (*string)(nil), result.StringPtrValue; expected != actual {
				t.Errorf("Expected nil string pointer value but got '%s'", *actual)
			}
			if expected, actual := 0, result.IntValue; expected != actual {
				t.Errorf("Expected int value %d but got %d", expected, actual)
			}
			if expected, actual := (*int)(nil), result.IntPtrValue; expected != actual {
				t.Errorf("Expected nil int pointer value but got %d", *actual)
			}
			if actual := result.BoolValue; actual {
				t.Errorf("Expected bool value to be false but wasn’t")
			}
			if actual := result.BoolPtrValue; actual != nil {
				t.Errorf("Expected bool pointer value to be nil but wasn’t")
			}
			if expected, actual := (time.Time{}), result.TimeValue; expected != actual {
				t.Errorf("Expected time value %v but got %v", expected, actual)
			}
			if expected, actual := (*time.Time)(nil), result.TimePtrValue; expected != actual {
				t.Errorf("Expected nil time pointer value but got %v", *actual)
			}
			if actual := result.NullTimeValue.Valid; actual {
				t.Errorf("Expected invalid nulltime value but was valid")
			}
			if expected, actual := "", result.StructValue.NestedStringValue; expected != actual {
				t.Errorf("Expected nested string value '%s' but got '%s'", expected, actual)
			}
			if expected, actual := (*string)(nil), result.StructValue.NestedStringPtrValue; expected != actual {
				t.Errorf("Expected nil nested string pointer value but got '%s'", *actual)
			}

			if actual := result.StructPtr; actual != nil {
				t.Errorf("Expected nested struct pointer to be nil but was instantiated")
			}
		})

		t.Run("maps successive structs without sharing pointers", func(t *testing.T) {
			ss := For((*TestStruct)(nil), "")

			mockQuery := fmt.Sprintf("some query")

			string1 := "string 1"
			string2 := "string 2"

			dbMock.ExpectQuery(mockQuery).WillReturnRows(
				sqlmock.NewRows([]string{
					"string_ptr_value",
				}).
					AddRow(&string1).
					AddRow(&string2),
			)

			rows, err := db.Query(mockQuery)
			if err != nil {
				t.Fatalf("Error executing query: %v", err)
			}

			if !rows.Next() {
				t.Fatalf("Expected a first row but got none")
			}

			var result1 TestStruct
			err = ss.Scan(rows, &result1)
			if err != nil {
				t.Fatalf("Expected success scanning first resut but got error: %v", err)
			}

			if !rows.Next() {
				t.Fatalf("Expected a second row but got none")
			}

			var result2 TestStruct
			err = ss.Scan(rows, &result2)
			if err != nil {
				t.Fatalf("Expected success scanning second result but got error: %v", err)
			}

			if expected, actual := string1, result1.StringPtrValue; actual == nil || expected != *actual {
				t.Errorf("Expected first result to point to '%s' but was %v", expected, actual)
			}
			if expected, actual := string2, result2.StringPtrValue; actual == nil || expected != *actual {
				t.Errorf("Expected second result to point to '%s' but was %v", expected, actual)
			}
		})

		t.Run("panics when no destination field exists to receive a value", func(t *testing.T) {
			ss := For((*TestStruct)(nil), "prefix")

			mockQuery := fmt.Sprintf("some query")

			dbMock.ExpectQuery(mockQuery).WillReturnRows(
				sqlmock.NewRows([]string{
					"prefix.nonexistent_field",
				}).AddRow(
					"string value",
				),
			)

			rows, err := db.Query(mockQuery)
			if err != nil {
				t.Fatalf("Error executing query: %v", err)
			}

			if !rows.Next() {
				t.Fatalf("Expected one row but got none")
			}

			var result TestStruct
			var panicValue any

			func() {
				defer func() {
					if r := recover(); r != nil {
						panicValue = r
					}
				}()

				_ = ss.Scan(rows, &result)
			}()

			if panicValue == nil {
				t.Fatalf("Expected a panic but returned from Scan")
			}

			switch v := panicValue.(type) {
			case string:
				if expected, actual := "no destination field for 'prefix.nonexistent_field'", v; expected != actual {
					t.Errorf("Expected panic with '%s' but was '%s'", expected, actual)
				}

			default:
				t.Fatalf("Panicked with unexpected value %v", panicValue)
			}
		})

		t.Run("when ignoring nonexistent destination fields", func(t *testing.T) {
			IgnoreNonexistentFields(true)

			t.Cleanup(func() {
				IgnoreNonexistentFields(false)
			})

			t.Run("does not panic", func(t *testing.T) {
				ss := For((*TestStruct)(nil), "prefix")

				mockQuery := fmt.Sprintf("some query")

				dbMock.ExpectQuery(mockQuery).WillReturnRows(
					sqlmock.NewRows([]string{
						"prefix.nonexistent_field",
					}).AddRow(
						"string value",
					),
				)

				rows, err := db.Query(mockQuery)
				if err != nil {
					t.Fatalf("Error executing query: %v", err)
				}

				if !rows.Next() {
					t.Fatalf("Expected one row but got none")
				}

				var result TestStruct
				var panicValue any

				func() {
					defer func() {
						if r := recover(); r != nil {
							panicValue = r
						}
					}()

					err = ss.Scan(rows, &result)
				}()

				if panicValue != nil {
					t.Fatalf("Expected success but panicked with: %v", panicValue)
				}
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}
			})
		})
	})
}
