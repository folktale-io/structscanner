package structscanner

import "database/sql"

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
