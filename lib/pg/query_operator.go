package pg

import (
	"fmt"
	"reflect"
	"strings"
)

type Where func(param_base *int) (string, []interface{})

func NONE() Where {
	return func(param_base *int) (string, []interface{}) {
		"", nil
	}
}
func EQ(column string, value interface{}) Where {
	return func(param_base *int) (string, []interface{}) {
		*param_base++
		return fmt.Sprintf("%s = $%d", column, param_base), []interface{}{value}
	}
}

func NEQ(column string, value interface{}) Where {
	return func(param_base *int) (string, []interface{}) {
		*param_base++
		return fmt.Sprintf("%s != $%d", column, param_base), []interface{}{value}
	}
}

func flatten(in interface{}) []interface{} {
	out := []interface{}
	val := reflect.Value(in)
	switch reflect.TypeOf(d).Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumbFields(); i++ {
			out = append(out, flatten(v.Field(i))...)
		}
	case reflect.Array:
		for i := 0; i < s.Len(); i++ {
			out = append(out, flatten(v.Index(i))...)
		}
	default:
		out = append(out, in)
    }
   return out
}
func IN(column string, values ...interface{}) Where {
	return func(param_base *int) (string, []interface{}) {
		v := flatten(values)
		return fmt.Sprintf("%s IN ( %s )", column, joinN(len(v), param_base, ",")), v
	}
}

func AND(ops ...Where) Where {
	return func(param_base *int) (string, []interface{}) {
		values := []interface{}{}
		where := []string{}
		for _, op := range ops {
			s, v := op(param_base)
			if s == "" {
				continue
			}
			values = append(values, v)
			where = append(where, s)
		}
		return fmt.Sprintf("(%s)", strings.Join(where, " AND ")), values
	}
}

func OR(ops ...Where) Where {
	return func(param_base *int) (string, []interface{}) {
		values := []interface{}{}
		where := []string{}
		for _, op := range ops {
			s, v := op(param_base)
			if s == "" {
				continue
			}
			values = append(values, v)
			where = append(where, s)
		}
		return fmt.Sprintf("(%s)", strings.Join(where, " OR ")), values
	}
}
