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

func (m *Product) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.ProductID = string(buff)
	return nil
}

func (m *Product) Value() (driver.Value, error) {
	return m.ProductID, nil
}

func (m *Product) GetIdentifier() interface{} {
	return m.ProductID
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
func (s *TestStore) selectProduct(ctx context.Context, config *queryProductConfig, withRow func(*Product)) error {
	base := 0
	placeholders := func(base *int, length int) []interface{} {
		arr := make([]interface{}, length)
		for i := range arr {
			*base++
			arr[i] = fmt.Sprintf("$%d", *base)
		}
		return arr
	}
	where, vals := pg.GetWhereClause(config.filter, &base)
	params := append([]interface{}{where}, placeholders(&base, 2)...)
	stmt, err := s.conn.PrepareContext(ctx, fmt.Sprintf(` 
	SELECT "product_id", "product_name", "product_config" 
	FROM "tbl_product"
	%s
	LIMIT %s OFFSET %s`, params...))
	if err != nil {
		return err
	}
	vals = append(vals, config.limit, config.start)
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		row := new(Product)
		err := cursor.Scan(&row.ProductID, &row.ProductName, &row.Config)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
