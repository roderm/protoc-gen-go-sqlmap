package pg

import (
	"fmt"
	"strings"
)

type inserVals struct {
	values [][]interface{}
}
func NewInsert() *inserVals {
	return new(inserVals)
}
func (v *inserVals) Add(values ...[]interface{}) error {
	v.values = append(v.values, values...)
}

func (v *inserVals) Values() []interface{} {
	values := []interface{}
	for _, r := range v.values {
		values = append(values, r...)
	}
	return values
}
func (v *inserVals) String() string {
	base := 1;
	mRows := []string{}
	for _, r := range v.values {
		mRows = append(mRows, 
			fmt.Sprintf("(%s)", joinN(len(r), *base, ", ")),
		)
	}
	return strings.Join(mRows, ", ")
}
