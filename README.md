# structscanner

`structscanner` is a simple library to make going from database queries to structs easier, while retaining the flexibility of joins and mapping using struct tags.

It has the following features:

* Works with `database/sql` without extra dependencies.
* Mapping based on `db:` struct tags—only fields with a `db:` tag are mapped.
* Mapping of NULLs to zero values—this makes handling outer joins much more practical without having to create alternate, nullable versions of your structs or peppering your queries with `IFNULL` or `COALESCE`.
* Mapping using an optional prefix at query time—this allows structs to be mapped more easily when table aliases are being used with column names (such as with the [`columnsWithAlias`](https://github.com/Go-SQL-Driver/MySQL/#columnswithalias) option with `go-sql-driver/mysql`).
* Convenient querying for single structs or slices of structs.
* Thread-safe caching of reflection metadata.

## Example usage

When passed a single struct as the destination, `Select` will populate it with the results of the first row, or return `sql.ErrNoRows` if there is none. 

```go
var result struct {
	AppleCount int `db:"fruit_count"`
	TotalFruit int `db:"total_fruit"`
}

err := structscanner.Select(db &result, "", `
    SELECT
        SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS fruit_count,
        COUNT(1) AS total_fruit
    FROM fruit_basket
`, "apple")
```

Often, you will have a struct with fields already tagged. When querying for such a struct, a problem that can arise is that the columns are returned with a table alias, which is missing from your struct tags. You can provide a prefix to `Select` in order to correctly map these to your existing struct:

```go
type Fruit struct {
	ID     uint64 `db:"id"`
	Name   string `db:"name"`
	Colour string `db:"colour"`
}

var result Fruit

err := structscanner.Select(db, &result, "f", `
    SELECT f.* FROM fruits f WHERE id = ?
`, fruitID)
``` 

When passed a slice as the destination, `Select` will populate it with all rows that are returned from the query. If there are no rows, the destination slice will be empty.

Querying for a collection of related entities:

```go
var result []struct {
	O Organisation `db:"o"`
	P Person       `db:"p"`
}

err := structscanner.Select(db, &result, "", `
    SELECT
        o.*,
        p.*
    FROM organisations o
    JOIN people p ON p.organisation_id = o.id
    WHERE o.id = ?
`, organisationID)
```

Querying for a collection of related entities where one side might be missing:

```go
var result []struct {
	Individual Person `db:"i"`
	Sibling    Person `db:"s"`
}

err := structscanner.Select(db, &result, "", `
    SELECT
        i.*,
        s.*
    FROM people i
    LEFT JOIN people s ON i.sibling_id = s.id
`)
```

In this case, if an individual doesn’t have a sibling, the `LEFT JOIN` will produce NULLs. As this sort of usage is quite common, the fields of `Sibling` are set to their zero values rather than this being treated as an error condition.
