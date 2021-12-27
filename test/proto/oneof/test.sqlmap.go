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
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Hardware) error
	cb           []func(*Hardware)
	rows         map[interface{}]*Hardware
}

type HardwareOption func(*queryHardwareConfig)

func HardwarePaging(page, length int) HardwareOption {
	return func(config *queryHardwareConfig) {
		config.start = length * page
		config.limit = length
	}
}
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
		limit:  1000,
		rows:   make(map[interface{}]*Hardware),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectHardware(ctx, config, func(row *Hardware) {
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
func (s *TestStore) selectHardware(ctx context.Context, config *queryHardwareConfig, withRow func(*Hardware)) error {
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
	SELECT "product_id", "product_serial" 
	FROM "tbl_hardware"
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
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Product) error
	cb           []func(*Product)
	rows         map[interface{}]*Product

	loadHardware bool
	optsHardware []HardwareOption

	loadService bool
	optsService []ServiceOption

	loadSoftware bool
	optsSoftware []SoftwareOption
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
func ProductWithService(opts ...ServiceOption) ProductOption {
	return func(config *queryProductConfig) {
		mapService := make(map[interface{}]*Product)
		config.loadService = true
		config.optsService = opts
		config.cb = append(config.cb, func(row *Product) {

			mapService[row.GetProductID()] = row

		})
		config.optsService = append(config.optsService,
			ServiceOnRow(func(row *Service) {

				// one-to-one
				item, ok := mapService[row.ProductID]
				if ok && item != nil {
					if config.rows[item.ProductID] != nil {
						config.rows[item.ProductID].Type = &Product_Service{Service: row}
					}
				}

			}),
			ServiceFilter(pg.INCallabel("product_id", func() []interface{} {
				ids := []interface{}{}
				for id := range mapService {
					ids = append(ids, id)
				}
				return ids
			})),
		)
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

	if config.loadHardware {
		// github.com/roderm/protoc-gen-go-sqlmap/test/proto/oneof

		_, err = s.Hardware(ctx, config.optsHardware...)

	}
	if err != nil {
		return config.rows, err
	}

	if config.loadService {
		// github.com/roderm/protoc-gen-go-sqlmap/test/proto/oneof

		_, err = s.Service(ctx, config.optsService...)

	}
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
	SELECT "product_id", "product_name", "product_type" 
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
		err := cursor.Scan(&row.ProductID, &row.ProductName, &row.ProductType)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}

func (m *Service) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.ProductID = string(buff)
	return nil
}

func (m *Service) Value() (driver.Value, error) {
	return m.ProductID, nil
}

func (m *Service) GetIdentifier() interface{} {
	return m.ProductID
}

type queryServiceConfig struct {
	Store        *TestStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Service) error
	cb           []func(*Service)
	rows         map[interface{}]*Service
}

type ServiceOption func(*queryServiceConfig)

func ServicePaging(page, length int) ServiceOption {
	return func(config *queryServiceConfig) {
		config.start = length * page
		config.limit = length
	}
}
func ServiceFilter(filter pg.Where) ServiceOption {
	return func(config *queryServiceConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func ServiceOnRow(cb func(*Service)) ServiceOption {
	return func(s *queryServiceConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Service(ctx context.Context, opts ...ServiceOption) (map[interface{}]*Service, error) {
	config := &queryServiceConfig{
		Store:  s,
		filter: pg.NONE(),
		limit:  1000,
		rows:   make(map[interface{}]*Service),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectService(ctx, config, func(row *Service) {
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
func (s *TestStore) selectService(ctx context.Context, config *queryServiceConfig, withRow func(*Service)) error {
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
	SELECT "product_id" 
	FROM "tbl_service"
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
		row := new(Service)
		err := cursor.Scan(&row.ProductID)
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
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Software) error
	cb           []func(*Software)
	rows         map[interface{}]*Software
}

type SoftwareOption func(*querySoftwareConfig)

func SoftwarePaging(page, length int) SoftwareOption {
	return func(config *querySoftwareConfig) {
		config.start = length * page
		config.limit = length
	}
}
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
		limit:  1000,
		rows:   make(map[interface{}]*Software),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectSoftware(ctx, config, func(row *Software) {
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
func (s *TestStore) selectSoftware(ctx context.Context, config *querySoftwareConfig, withRow func(*Software)) error {
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
	SELECT "product_id", "product_version" 
	FROM "tbl_software"
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
		row := new(Software)
		err := cursor.Scan(&row.ProductID, &row.Version)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
