// Package query is the runtime half of the query writer: the small amount of
// logic that would otherwise be duplicated into every generated file.
//
// Generated code stays dialect-agnostic -- it emits column and table names and
// leaves placeholder syntax, field-mask interpretation and row batching to the
// helpers here, so one generated file serves PostgreSQL, MySQL and SQLite.
package query

import (
	"context"
	"database/sql"
	"slices"
	"strconv"
	"strings"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// DB is the subset of *sql.DB that generated queries need, so a *sql.Tx or a
// pooling wrapper works just as well.
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Scanner is implemented by both *sql.Row and *sql.Rows.
type Scanner interface {
	Scan(dest ...any) error
}

// Dialect selects the placeholder syntax for a connection.
type Dialect string

const (
	Postgres Dialect = "postgres"
	MySQL    Dialect = "mysql"
	SQLite   Dialect = "sqlite3"
)

// Placeholder renders the n-th (1-based) bind placeholder for the dialect.
func (d Dialect) Placeholder(n int) string {
	if d == Postgres {
		return "$" + strconv.Itoa(n)
	}
	return "?"
}

// Conn pairs a database handle with the dialect its placeholders use.
type Conn struct {
	DB      DB
	Dialect Dialect
}

// Mask is a parsed google.protobuf.FieldMask: a tree of selected field names,
// where each node may carry a sub-mask for a nested message.
//
// A nil *Mask selects everything, which makes "no mask supplied" and "no
// restriction below this point" the same value and keeps callers free of nil
// checks.
type Mask struct {
	fields map[string]*Mask
}

// FromFieldMask converts a protobuf FieldMask into a Mask. A nil or empty mask
// selects everything.
func FromFieldMask(fm *fieldmaskpb.FieldMask) *Mask {
	if fm == nil || len(fm.GetPaths()) == 0 {
		return nil
	}
	return NewMask(fm.GetPaths()...)
}

// NewMask builds a Mask from dotted field paths, e.g. "name", "books.title".
// No paths selects everything.
func NewMask(paths ...string) *Mask {
	if len(paths) == 0 {
		return nil
	}
	root := &Mask{fields: map[string]*Mask{}}
	for _, path := range paths {
		node := root
		for part := range strings.SplitSeq(path, ".") {
			if part == "" {
				continue
			}
			// A nil fields map means an earlier, shorter path already selected
			// everything below this node, so a narrower path adds nothing --
			// and descending further would write to that nil map.
			if node.fields == nil {
				node = nil
				break
			}
			next, ok := node.fields[part]
			if !ok || next == nil {
				next = &Mask{fields: map[string]*Mask{}}
				node.fields[part] = next
			}
			node = next
		}
		// The path ends here, so it selects the whole subtree below it: drop
		// any narrower sub-mask a longer sibling had installed, so that
		// ["books.title", "books"] selects all of books, like ["books"] alone.
		if node != nil && node != root {
			node.fields = nil
		}
	}
	return root
}

// Has reports whether a column is selected by the mask. An absent mask
// selects every column.
func (m *Mask) Has(name string) bool {
	if m == nil || m.fields == nil {
		return true
	}
	_, ok := m.fields[name]
	return ok
}

// HasRelation reports whether a relation is selected by the mask.
//
// Unlike Has, an absent mask selects *no* relations. Eager loading has to be
// asked for: "everything" would otherwise walk the whole object graph, which
// never terminates once two tables reference each other. It also means the
// depth of loading is exactly the depth spelled out in the mask, since the
// sub-mask below a leaf path is itself absent.
func (m *Mask) HasRelation(name string) bool {
	if m == nil || m.fields == nil {
		return false
	}
	_, ok := m.fields[name]
	return ok
}

// Sub returns the mask that applies below name. It is nil (meaning "select
// everything") when name was selected without a narrower sub-path.
func (m *Mask) Sub(name string) *Mask {
	if m == nil || m.fields == nil {
		return nil
	}
	sub, ok := m.fields[name]
	if !ok || sub == nil || sub.fields == nil {
		return nil
	}
	return sub
}

// Cond is a SQL fragment and its bind arguments. Placeholders are written as
// "?" and renumbered per dialect when the statement is assembled, so callers
// never deal with $1/$2 themselves.
type Cond struct {
	SQL  string
	Args []any
}

// In builds a `col IN (?, ?, ...)` condition. It returns false when values is
// empty, since `IN ()` is not valid SQL and the caller should skip the query
// entirely rather than run one that cannot match.
func In(col string, values []any) (Cond, bool) {
	if len(values) == 0 {
		return Cond{}, false
	}
	var b strings.Builder
	b.WriteString(col)
	b.WriteString(" IN (")
	for i := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("?")
	}
	b.WriteString(")")
	return Cond{SQL: b.String(), Args: values}, true
}

// Select assembles and runs a SELECT, renumbering the "?" placeholders in the
// conditions to the dialect's syntax and joining them with AND.
func (c Conn) Select(ctx context.Context, table string, cols []string, conds []Cond) (*sql.Rows, error) {
	var b strings.Builder
	b.WriteString("SELECT ")
	for i, col := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(col)
	}
	b.WriteString(" FROM ")
	b.WriteString(table)

	var args []any
	for i, cond := range conds {
		if i == 0 {
			b.WriteString(" WHERE ")
		} else {
			b.WriteString(" AND ")
		}
		b.WriteString(c.renumber(cond.SQL, len(args)))
		args = append(args, cond.Args...)
	}
	return c.DB.QueryContext(ctx, b.String(), args...)
}

// renumber rewrites each "?" in sql to the dialect's placeholder, continuing
// from the count of arguments already bound by earlier conditions.
func (c Conn) renumber(sql string, bound int) string {
	if c.Dialect != Postgres {
		return sql
	}
	var b strings.Builder
	for _, r := range sql {
		if r == '?' {
			bound++
			b.WriteString(c.Dialect.Placeholder(bound))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Keys returns the keys of m, which is how a batch of loaded rows turns into
// the IN list for fetching their related rows.
func Keys[K comparable, V any](m map[K]V) []any {
	out := make([]any, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// Contains reports whether s holds v.
func Contains[T comparable](s []T, v T) bool {
	return slices.Contains(s, v)
}

// Key normalizes a join key so that a value read back from a driver and one
// read off a generated getter compare equal.
//
// Drivers are not consistent about the Go type they produce: the same column
// can arrive as []byte or string, or as any width of integer, depending on
// driver and dialect. Since one side of every stitch comes from a driver (the
// raw foreign-key value) and the other from a typed getter, both go through
// here before being used as a map key. A nil value returns nil, which callers
// treat as "no key" and skip.
func Key(v any) any {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		return string(t)
	case int:
		return int64(t)
	case int8:
		return int64(t)
	case int16:
		return int64(t)
	case int32:
		return int64(t)
	case uint:
		return int64(t)
	case uint8:
		return int64(t)
	case uint16:
		return int64(t)
	case uint32:
		return int64(t)
	case uint64:
		return int64(t)
	case float32:
		return float64(t)
	default:
		return v
	}
}
