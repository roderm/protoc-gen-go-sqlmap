package pg

import (
	"context"
	"strings"
)

type protoBase interface {
	GetColumns() []string
	GetTablename() string
	GetPK() string
	NewRow() interface{}
}

type Query struct {
	protoBase
	cols   []string
	filter Where
}

func (q *Query) Select(columns ...string) {

}

func (q *Query) Where(filter Where) {
	q.filter = filter
}

func (q *Query) String() (string, []interface{}) {
	base := 1
	cols := `SELECT ` + strings.Join(q.cols, ",")
	from := `FROM ` + q.GetTablename()
	where, vals := q.filter(&base)
	return cols + "\n" + from + "\n WHERE " + where, vals
}
func (q *Query) Go(ctx context.Context) error {
	return nil
}
