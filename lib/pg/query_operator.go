package pg

import (
	"fmt"
	"strings"
)

type Where func(param_base *int) (string, []interface{})

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

func IN(column string, values ...interface{}) Where {
	return func(param_base *int) (string, []interface{}) {
		return fmt.Sprintf("%s IN ( %s )", column, joinN(len(values), param_base, ",")), values
	}
}

func AND(ops ...Where) Where {
	return func(param_base *int) (string, []interface{}) {
		values := []interface{}{}
		where := []string{}
		for _, op := range ops {
			s, v := op(param_base)
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
			values = append(values, v)
			where = append(where, s)
		}
		return fmt.Sprintf("(%s)", strings.Join(where, " OR ")), values
	}
}
