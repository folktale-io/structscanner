package structscanner

var ignoreNonexistentFields = false

// IgnoreNonexistentFields sets whether missing struct fields (queried columns
// that have no mapped struct field to be stored in) trigger a panic, or are
// silently ignored.
func IgnoreNonexistentFields(ignore bool) {
	ignoreNonexistentFields = ignore
}
