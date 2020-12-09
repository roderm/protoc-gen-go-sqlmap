package pg

import (
	"fmt"
	"reflect"
	"strings"
)

// Where is a function type that is used to create a WHERE clause and
// the values to use in an SQL-Query.
type Where func(paramBase *int) (string, []interface{})

// GetWhereClause builds the WHERE clause from any of Where-types
func GetWhereClause(w Where) (string, []interface{}) {
	var i *int = new(int)
	*i = 0
	str, vals := w(i)
	if len(str) > 0 {
		return "WHERE " + str, vals
	}
	return "", []interface{}{}
}

// NONE for using non filtered input
func NONE() Where {
	return func(paramBase *int) (string, []interface{}) {
		return "", nil
	}
}

// EQ compares values in an SQL-Statement with "="-operator
func EQ(column string, value interface{}) Where {
	return func(paramBase *int) (string, []interface{}) {
		*paramBase++
		return fmt.Sprintf("\"%s\" = $%d", column, *paramBase), []interface{}{value}
	}
}

// NEQ compares values in an SQL-Statement with "!="-operator
func NEQ(column string, value interface{}) Where {
	return func(paramBase *int) (string, []interface{}) {
		*paramBase++
		return fmt.Sprintf("\"%s\" != $%d", column, *paramBase), []interface{}{value}
	}
}

func flatten(in interface{}) []interface{} {
	out := []interface{}{}
	val := reflect.ValueOf(in)
	switch reflect.TypeOf(in).Kind() {
	case reflect.Struct:
		for i := 0; i < reflect.TypeOf(in).NumField(); i++ {
			out = append(out, flatten(val.Field(i))...)
		}
	case reflect.Array:
		for i := 0; i < val.Len(); i++ {
			out = append(out, flatten(val.Index(i))...)
		}
	default:
		out = append(out, in)
	}
	return out
}

// IN compares values in an SQL-Statement with "IN (?,?,?,...)"-operator
func IN(column string, values ...interface{}) Where {
	return func(paramBase *int) (string, []interface{}) {
		v := flatten(values)
		return fmt.Sprintf("\"%s\" IN ( %s )", column, joinN(len(v), paramBase, ",")), v
	}
}

// AND joins two or more Where-types with (cond1 AND cond2)
func AND(ops ...Where) Where {
	return func(paramBase *int) (string, []interface{}) {
		values := []interface{}{}
		where := []string{}
		for _, op := range ops {
			s, v := op(paramBase)
			if s == "" {
				continue
			}
			values = append(values, v...)
			where = append(where, s)
		}
		return fmt.Sprintf("(%s)", strings.Join(where, " AND ")), values
	}
}

// OR joins two or more Where-types with (cond1 OR cond2)
func OR(ops ...Where) Where {
	return func(paramBase *int) (string, []interface{}) {
		values := []interface{}{}
		where := []string{}
		for _, op := range ops {
			s, v := op(paramBase)
			if s == "" {
				continue
			}
			values = append(values, v...)
			where = append(where, s)
		}
		return fmt.Sprintf("(%s)", strings.Join(where, " OR ")), values
	}
}
