package structscanner

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestStructScanner(t *testing.T) {

	db, dbMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error setting up mock DB: %v", err)
	}
	defer db.Close()

	type TestStruct struct {
		StringValue    string       `db:"string_value"`
		StringPtrValue *string      `db:"string_ptr_value"`
		IntValue       int          `db:"int_value"`
		IntPtrValue    *int         `db:"int_ptr_value"`
		BoolValue      bool         `db:"bool_value"`
		BoolPtrValue   *bool        `db:"bool_ptr_value"`
		TimeValue      time.Time    `db:"time_value"`
		TimePtrValue   *time.Time   `db:"time_ptr_value"`
		NullTimeValue  sql.NullTime `db:"null_time_value"`
		StructValue    struct {
			NestedStringValue    string  `db:"nested_string_value"`
			NestedStringPtrValue *string `db:"nested_string_ptr_value"`
		} `db:"struct_value"`
	}

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
				"nested string value",
				"nested string ptr value",
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
		if expected, actual := "nested string value", result.StructValue.NestedStringValue; expected != actual {
			t.Errorf("Expected nested string value '%s' but got '%s'", expected, actual)
		}
		if expected, actual := "nested string ptr value", result.StructValue.NestedStringPtrValue; expected != *actual {
			t.Errorf("Expected nested string pointer value '%s' but got '%s'", expected, *actual)
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
	})
}
