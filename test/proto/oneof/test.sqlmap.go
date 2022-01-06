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

func (s *TestStore) GetHardwareSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "product_id", "product_serial"
		FROM "tbl_hardware"
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

func (s *TestStore) selectHardware(ctx context.Context, config *queryHardwareConfig, withRow func(*Hardware)) error {
	query, vals := s.GetHardwareSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectHardware': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectHardware' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Hardware{}
		err := cursor.Scan(&row.ProductID, &row.Serial)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}

type queryProductConfig struct {
	Store        *TestStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Product) error
	cb           []func(*Product)
	rows         map[interface{}]*Product
	loadSoftware bool
	optsSoftware []SoftwareOption
	loadHardware bool
	optsHardware []HardwareOption
	loadService  bool
	optsService  []ServiceOption
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

func ProductWithSoftware(opts ...SoftwareOption) ProductOption {
	return func(config *queryProductConfig) {
		config.loadSoftware = true
		parent := make(map[interface{}][]*Product)
		config.cb = append(config.cb, func(row *Product) {
			child_key := row.ProductID
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsSoftware = append(opts,
			SoftwareFilter(
				pg.INCallabel("product_id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			SoftwareOnRow(func(row *Software) {
				children := parent[row.ProductID]
				for _, c := range children {
					c.Type = &Product_Software{Software: row}
				}
			}),
		)
	}
}

func ProductWithHardware(opts ...HardwareOption) ProductOption {
	return func(config *queryProductConfig) {
		config.loadHardware = true
		parent := make(map[interface{}][]*Product)
		config.cb = append(config.cb, func(row *Product) {
			child_key := row.ProductID
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsHardware = append(opts,
			HardwareFilter(
				pg.INCallabel("product_id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			HardwareOnRow(func(row *Hardware) {
				children := parent[row.ProductID]
				for _, c := range children {
					c.Type = &Product_Hardware{Hardware: row}
				}
			}),
		)
	}
}

func ProductWithService(opts ...ServiceOption) ProductOption {
	return func(config *queryProductConfig) {
		config.loadService = true
		parent := make(map[interface{}][]*Product)
		config.cb = append(config.cb, func(row *Product) {
			child_key := row.ProductID
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsService = append(opts,
			ServiceFilter(
				pg.INCallabel("product_id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			ServiceOnRow(func(row *Service) {
				children := parent[row.ProductID]
				for _, c := range children {
					c.Type = &Product_Service{Service: row}
				}
			}),
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

func (s *TestStore) GetProductSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "product_id", "product_name", "product_type"
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
		err := cursor.Scan(&row.ProductID, &row.ProductName, &row.ProductType)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
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

func (s *TestStore) GetServiceSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "product_id"
		FROM "tbl_service"
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

func (s *TestStore) selectService(ctx context.Context, config *queryServiceConfig, withRow func(*Service)) error {
	query, vals := s.GetServiceSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectService': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectService' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Service{}
		err := cursor.Scan(&row.ProductID)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
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

func (s *TestStore) GetSoftwareSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "product_id", "product_version"
		FROM "tbl_software"
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

func (s *TestStore) selectSoftware(ctx context.Context, config *querySoftwareConfig, withRow func(*Software)) error {
	query, vals := s.GetSoftwareSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectSoftware': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectSoftware' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Software{}
		err := cursor.Scan(&row.ProductID, &row.Version)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
