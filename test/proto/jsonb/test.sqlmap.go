package jsonb

import (
	context "context"
	sql "database/sql"
	driver "database/sql/driver"
	json "encoding/json"
	fmt "fmt"
	pg "github.com/roderm/protoc-gen-go-sqlmap/lib/go/pg"
)

var _ = fmt.Sprintf
var _ = context.TODO
var _ = pg.NONE
var _ = sql.Open
var _ = driver.IsValue
var _ = json.Valid

type TestStore struct {
	conn *sql.DB
}

func NewTestStore(conn *sql.DB) *TestStore {
	return &TestStore{conn}
}

type queryProductConfig struct {
	Store        *TestStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Product) error
	cb           []func(*Product)
	rows         map[interface{}]*Product
}

type ProductOption func(*queryProductConfig)

func ProductPaging(page, length int) ProductOption {
	return func(config *queryProductConfig) {
		config.start = length * page
		config.limit = length
	}
}
func ProductFilter(filter pg.Where) ProductOption {
	return func(config *queryProductConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func ProductOnRow(cb func(*Product)) ProductOption {
	return func(s *queryProductConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Product(ctx context.Context, opts ...ProductOption) (map[interface{}]*Product, error) {
	config := &queryProductConfig{
		Store:  s,
		filter: pg.NONE(),
		limit:  1000,
		rows:   make(map[interface{}]*Product),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectProduct(ctx, config, func(row *Product) {
		config.rows[row.ProductID] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}

	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *TestStore) GetProductSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "product_id", "product_name", "product_config"
		FROM "tbl_product"
		%s`, where)

	if limit > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nLIMIT $%d", base)
		vals = append(vals, limit)
	}
	if start > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nOFFSET $%d", base)
		vals = append(vals, start)
	}
	return tpl, vals
}

func (s *TestStore) selectProduct(ctx context.Context, config *queryProductConfig, withRow func(*Product)) error {
	query, vals := s.GetProductSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectProduct': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectProduct' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Product{}
		err := cursor.Scan(&row.ProductID, &row.ProductName)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
