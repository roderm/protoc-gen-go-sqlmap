package oneof

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

type HardwareList map[interface{}]*Hardware

func (m *Hardware) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"product_id": m.ProductID,
	}
	return pk
}
func (m *Hardware) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Hardware' failed: %s", string(buff), err)
	}
	return nil
}

type queryHardwareConfig struct {
	Store        *TestStore
	filter       []squirrel.Sqlizer
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

func HardwareFilter(filter ...squirrel.Sqlizer) HardwareOption {
	return func(config *queryHardwareConfig) {
		config.filter = append(config.filter, filter...)
	}
}

func HardwareOnRow(cb func(*Hardware)) HardwareOption {
	return func(s *queryHardwareConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Hardware(ctx context.Context, opts ...HardwareOption) (HardwareList, error) {
	config := &queryHardwareConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(HardwareList),
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

func (s *TestStore) GetHardwareSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"product_id", "product_serial"`).
		From("\"tbl_hardware\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TestStore) selectHardware(ctx context.Context, config *queryHardwareConfig, withRow func(*Hardware)) error {
	query := s.GetHardwareSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectHardware': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Hardware{}
	for cursor.Next() {
		row := new(Hardware)
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
	loadService  bool
	optsService  []ServiceOption
	loadSoftware bool
	optsSoftware []SoftwareOption
	loadHardware bool
	optsHardware []HardwareOption
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
				squirrel1.EqCall{"product_id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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
				squirrel1.EqCall{"product_id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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
				squirrel1.EqCall{"product_id": func() interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}},
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
	if config.loadHardware {
		_, err = s.Hardware(ctx, config.optsHardware...)
		if err != nil {
			return config.rows, err
		}
	}
	if config.loadService {
		_, err = s.Service(ctx, config.optsService...)
		if err != nil {
			return config.rows, err
		}
	}
	if config.loadSoftware {
		_, err = s.Software(ctx, config.optsSoftware...)
		if err != nil {
			return config.rows, err
		}
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
		Select(`"product_id", "product_name", "product_type"`).
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

type ServiceList map[interface{}]*Service

func (m *Service) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"product_id": m.ProductID,
	}
	return pk
}
func (m *Service) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Service' failed: %s", string(buff), err)
	}
	return nil
}

type queryServiceConfig struct {
	Store        *TestStore
	filter       []squirrel.Sqlizer
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

func ServiceFilter(filter ...squirrel.Sqlizer) ServiceOption {
	return func(config *queryServiceConfig) {
		config.filter = append(config.filter, filter...)
	}
}

func ServiceOnRow(cb func(*Service)) ServiceOption {
	return func(s *queryServiceConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Service(ctx context.Context, opts ...ServiceOption) (ServiceList, error) {
	config := &queryServiceConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(ServiceList),
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

func (s *TestStore) GetServiceSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"product_id"`).
		From("\"tbl_service\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TestStore) selectService(ctx context.Context, config *queryServiceConfig, withRow func(*Service)) error {
	query := s.GetServiceSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectService': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Service{}
	for cursor.Next() {
		row := new(Service)
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

type SoftwareList map[interface{}]*Software

func (m *Software) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"product_id": m.ProductID,
	}
	return pk
}
func (m *Software) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Software' failed: %s", string(buff), err)
	}
	return nil
}

type querySoftwareConfig struct {
	Store        *TestStore
	filter       []squirrel.Sqlizer
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

func SoftwareFilter(filter ...squirrel.Sqlizer) SoftwareOption {
	return func(config *querySoftwareConfig) {
		config.filter = append(config.filter, filter...)
	}
}

func SoftwareOnRow(cb func(*Software)) SoftwareOption {
	return func(s *querySoftwareConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Software(ctx context.Context, opts ...SoftwareOption) (SoftwareList, error) {
	config := &querySoftwareConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(SoftwareList),
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

func (s *TestStore) GetSoftwareSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"product_id", "product_version"`).
		From("\"tbl_software\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TestStore) selectSoftware(ctx context.Context, config *querySoftwareConfig, withRow func(*Software)) error {
	query := s.GetSoftwareSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectSoftware': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Software{}
	for cursor.Next() {
		row := new(Software)
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
