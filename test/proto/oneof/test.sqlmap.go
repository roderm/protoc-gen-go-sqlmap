package oneof

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

func (m *Hardware) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.ProductID = string(buff)
	return nil
}

func (m *Hardware) Value() (driver.Value, error) {
	return m.ProductID, nil
}

func (m *Hardware) GetIdentifier() interface{} {
	return m.ProductID
}

type queryHardwareConfig struct {
	Store        *TestStore
	filter       pg.Where
	beforeReturn []func(map[interface{}]*Hardware) error
	cb           []func(*Hardware)
	rows         map[interface{}]*Hardware
}

type HardwareOption func(*queryHardwareConfig)

func HardwareFilter(filter pg.Where) HardwareOption {
	return func(config *queryHardwareConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func HardwareOnRow(cb func(*Hardware)) HardwareOption {
	return func(s *queryHardwareConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Hardware(ctx context.Context, opts ...HardwareOption) (map[interface{}]*Hardware, error) {
	config := &queryHardwareConfig{
		Store:  s,
		filter: pg.NONE(),
		rows:   make(map[interface{}]*Hardware),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectHardware(ctx, config.filter, func(row *Hardware) {
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
func (s *TestStore) selectHardware(ctx context.Context, filter pg.Where, withRow func(*Hardware)) error {
	where, vals := pg.GetWhereClause(filter, nil)
	stmt, err := s.conn.PrepareContext(ctx, ` 
	SELECT "product_id", "product_serial" 
	FROM "tbl_hardware"
	`+where)
	if err != nil {
		return err
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		row := new(Hardware)
		err := cursor.Scan(&row.ProductID, &row.Serial)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
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
	beforeReturn []func(map[interface{}]*Product) error
	cb           []func(*Product)
	rows         map[interface{}]*Product

	loadSoftware bool
	optsSoftware []SoftwareOption

	loadHardware bool
	optsHardware []HardwareOption
}

type ProductOption func(*queryProductConfig)

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

func ProductWithSoftware(opts ...SoftwareOption) ProductOption {
	return func(config *queryProductConfig) {
		mapSoftware := make(map[interface{}]*Product)
		config.loadSoftware = true
		config.optsSoftware = opts
		config.cb = append(config.cb, func(row *Product) {

			mapSoftware[row.GetProductID()] = row

		})
		config.optsSoftware = append(config.optsSoftware,
			SoftwareOnRow(func(row *Software) {

				// one-to-one
				item, ok := mapSoftware[row.ProductID]
				if ok && item != nil {
					if config.rows[item.ProductID] != nil {
						config.rows[item.ProductID].Type = &Product_Software{Software: row}
					}
				}

			}),
			SoftwareFilter(pg.INCallabel("product_id", func() []interface{} {
				ids := []interface{}{}
				for id := range mapSoftware {
					ids = append(ids, id)
				}
				return ids
			})),
		)
	}
}
func ProductWithHardware(opts ...HardwareOption) ProductOption {
	return func(config *queryProductConfig) {
		mapHardware := make(map[interface{}]*Product)
		config.loadHardware = true
		config.optsHardware = opts
		config.cb = append(config.cb, func(row *Product) {

			mapHardware[row.GetProductID()] = row

		})
		config.optsHardware = append(config.optsHardware,
			HardwareOnRow(func(row *Hardware) {

				// one-to-one
				item, ok := mapHardware[row.ProductID]
				if ok && item != nil {
					if config.rows[item.ProductID] != nil {
						config.rows[item.ProductID].Type = &Product_Hardware{Hardware: row}
					}
				}

			}),
			HardwareFilter(pg.INCallabel("product_id", func() []interface{} {
				ids := []interface{}{}
				for id := range mapHardware {
					ids = append(ids, id)
				}
				return ids
			})),
		)
	}
}

func (s *TestStore) Product(ctx context.Context, opts ...ProductOption) (map[interface{}]*Product, error) {
	config := &queryProductConfig{
		Store:  s,
		filter: pg.NONE(),
		rows:   make(map[interface{}]*Product),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectProduct(ctx, config.filter, func(row *Product) {
		config.rows[row.ProductID] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}

	if config.loadSoftware {
		// github.com/roderm/protoc-gen-go-sqlmap/test/proto/oneof

		_, err = s.Software(ctx, config.optsSoftware...)

	}
	if err != nil {
		return config.rows, err
	}

	if config.loadHardware {
		// github.com/roderm/protoc-gen-go-sqlmap/test/proto/oneof

		_, err = s.Hardware(ctx, config.optsHardware...)

	}
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
func (s *TestStore) selectProduct(ctx context.Context, filter pg.Where, withRow func(*Product)) error {
	where, vals := pg.GetWhereClause(filter, nil)
	stmt, err := s.conn.PrepareContext(ctx, ` 
	SELECT "product_id", "product_name", "product_type" 
	FROM "tbl_product"
	`+where)
	if err != nil {
		return err
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		row := new(Product)
		err := cursor.Scan(&row.ProductID, &row.ProductName, &row.ProductType)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}

func (m *Software) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.ProductID = string(buff)
	return nil
}

func (m *Software) Value() (driver.Value, error) {
	return m.ProductID, nil
}

func (m *Software) GetIdentifier() interface{} {
	return m.ProductID
}

type querySoftwareConfig struct {
	Store        *TestStore
	filter       pg.Where
	beforeReturn []func(map[interface{}]*Software) error
	cb           []func(*Software)
	rows         map[interface{}]*Software
}

type SoftwareOption func(*querySoftwareConfig)

func SoftwareFilter(filter pg.Where) SoftwareOption {
	return func(config *querySoftwareConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func SoftwareOnRow(cb func(*Software)) SoftwareOption {
	return func(s *querySoftwareConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Software(ctx context.Context, opts ...SoftwareOption) (map[interface{}]*Software, error) {
	config := &querySoftwareConfig{
		Store:  s,
		filter: pg.NONE(),
		rows:   make(map[interface{}]*Software),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectSoftware(ctx, config.filter, func(row *Software) {
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
func (s *TestStore) selectSoftware(ctx context.Context, filter pg.Where, withRow func(*Software)) error {
	where, vals := pg.GetWhereClause(filter, nil)
	stmt, err := s.conn.PrepareContext(ctx, ` 
	SELECT "product_id", "product_version" 
	FROM "tbl_software"
	`+where)
	if err != nil {
		return err
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		row := new(Software)
		err := cursor.Scan(&row.ProductID, &row.Version)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
