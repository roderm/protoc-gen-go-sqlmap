package jsonb

import (
	context "context"
	driver "database/sql/driver"
	json "encoding/json"
	fmt "fmt"
	squirrel "github.com/Masterminds/squirrel"
	sqlx "github.com/jmoiron/sqlx"
	squirrel1 "github.com/roderm/gotools/squirrel"
)

var _ = fmt.Sprintf
var _ = context.TODO
var _ = driver.IsValue
var _ = json.Valid
var _ = squirrel.Select
var _ = sqlx.Connect
var _ = squirrel1.EqCall{}

type TestStore struct {
	conn *sqlx.DB
}

func NewTestStore(conn *sqlx.DB) *TestStore {
	return &TestStore{conn}
}

type ProductList map[interface{}]*Product

func (m *Product) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"product_id": m.ProductID,
	}
	return pk
}
func (m *Product) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Product' failed: %s", string(buff), err)
	}
	return nil
}

type queryProductConfig struct {
	Store        *TestStore
	filter       []squirrel.Sqlizer
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

func ProductFilter(filter ...squirrel.Sqlizer) ProductOption {
	return func(config *queryProductConfig) {
		config.filter = append(config.filter, filter...)
	}
}

func ProductOnRow(cb func(*Product)) ProductOption {
	return func(s *queryProductConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Product(ctx context.Context, opts ...ProductOption) (ProductList, error) {
	config := &queryProductConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(ProductList),
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

func (s *TestStore) GetProductSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"product_id", "product_name", "product_config"`).
		From("\"tbl_product\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TestStore) selectProduct(ctx context.Context, config *queryProductConfig, withRow func(*Product)) error {
	query := s.GetProductSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectProduct': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Product{}
	for cursor.Next() {
		row := new(Product)
		err = cursor.StructScan(row)
		if err == nil {
			withRow(row)
			resultRows = append(resultRows, row)
		} else {
			return fmt.Errorf("sqlx.StructScan failed: %s", err)
		}

	}
	// err = sqlx.StructScan(cursor, &resultRows)
	// if err != nil {
	// 	return fmt.Errorf("sqlx.StructScan failed: %s", err)
	// }
	// for _, row := range resultRows {
	// 	withRow(row)
	// }
	return nil
}
