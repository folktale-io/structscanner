package structscanner

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
)

func TestSelect(t *testing.T) {

	db, dbMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Error setting up mock DB: %v", err)
	}
	defer db.Close()

	t.Run("maps results to single struct", func(t *testing.T) {

		var result struct {
			Value string `db:"value"`
		}

		query := `SELECT ? AS "prefix.value"`

		dbMock.
			ExpectQuery(query).
			WithArgs("some string").
			WillReturnRows(
				sqlmock.NewRows([]string{"prefix.value"}).
					AddRow("some string"),
			)

		err = Select(db, &result, "prefix", query, "some string")

		if err != nil {
			t.Fatalf("Expected query to succeed but got error: %v", err)
		}

		if expected, actual := "some string", result.Value; expected != actual {
			t.Errorf("Expected value '%s' but got '%s'", expected, actual)
		}
	})

	t.Run("returns ErrNoRows when destination is a single struct and there are no results", func(t *testing.T) {

		var result struct {
			Value string `db:"value"`
		}

		query := `SELECT TRUE AS "prefix.value" WHERE FALSE`

		dbMock.
			ExpectQuery(query).
			WillReturnRows(
				sqlmock.NewRows([]string{"prefix.value"}),
			)

		err := Select(db, &result, "prefix", query)

		if !errors.Is(err, sql.ErrNoRows) {
			if err == nil {
				t.Errorf("Expected an error but succeeded")
			} else {
				t.Errorf("Expected ErrNoRows but got error: %v", err)
			}
		}
	})

	t.Run("maps results to slice of structs", func(t *testing.T) {

		var result []struct {
			Value string `db:"value"`
		}

		query := `
			SELECT ? AS "prefix.value"
			UNION SELECT ? AS "prefix.value"
		`

		dbMock.
			ExpectQuery(query).
			WithArgs("some string 1", "some string 2").
			WillReturnRows(
				sqlmock.NewRows([]string{"prefix.value"}).
					AddRow("some string 1").
					AddRow("some string 2"),
			)

		err := Select(db, &result, "prefix", query, "some string 1", "some string 2")
		if err != nil {
			t.Fatalf("Expected success but got error: %v", err)
		}

		if expected, actual := 2, len(result); expected != actual {
			t.Fatalf("Expected %d results but got %d", expected, actual)
		}
		if expected, actual := "some string 1", result[0].Value; expected != actual {
			t.Errorf("Expected result 0 to have value '%s' but got '%s'", expected, actual)
		}
		if expected, actual := "some string 2", result[1].Value; expected != actual {
			t.Errorf("Expected result 1 to have value '%s' but got '%s'", expected, actual)
		}
	})

	t.Run("returns empty slice when destination is a slice and there are no results", func(t *testing.T) {

		var result []struct {
			Value string `db:"value"`
		}

		query := `SELECT TRUE AS "prefix.value" WHERE FALSE`

		dbMock.
			ExpectQuery(query).
			WillReturnRows(
				sqlmock.NewRows([]string{"prefix.value"}),
			)

		err := Select(db, &result, "prefix", query)

		if err != nil {
			t.Fatalf("Expected success but got error: %v", err)
		}

		if expected, actual := 0, len(result); expected != actual {
			t.Errorf("Expected no results but got %d", actual)
		}
	})
}
