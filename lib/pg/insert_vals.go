package pg

import (
	"fmt"
	"strings"
)

type inserVals struct {
	values [][]interface{}
}

// NewInsert creates a store for rows that
// will be inserted with an SQL-statement
func NewInsert() *inserVals {
	return new(inserVals)
}

// Add appends one or more rows to the insert store
func (v *inserVals) Add(values ...[]interface{}) error {
	v.values = append(v.values, values...)
	return nil
}

// Values returns all stored rows in a single-dimension array to use with the sql package
func (v *inserVals) Values() []interface{} {
	values := []interface{}{}
	for _, r := range v.values {
		values = append(values, r...)
	}
	return values
}

// String returns the VALUES-Clause with the size of given value-rows
func (v *inserVals) String() string {
	base := 1
	mRows := []string{}
	for _, r := range v.values {
		mRows = append(mRows, fmt.Sprintf("(%s)", joinN(len(r), &base, ", ")))
	}
	return strings.Join(mRows, ", ")
}
