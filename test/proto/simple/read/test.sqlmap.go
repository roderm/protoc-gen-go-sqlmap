package read

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

func (m *Employee) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.Id = string(buff)
	return nil
}

func (m *Employee) Value() (driver.Value, error) {
	return m.Id, nil
}

func (m *Employee) GetIdentifier() interface{} {
	return m.Id
}

type queryEmployeeConfig struct {
	Store        *TestStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Employee) error
	cb           []func(*Employee)
	rows         map[interface{}]*Employee
}

type EmployeeOption func(*queryEmployeeConfig)

func EmployeePaging(page, length int) EmployeeOption {
	return func(config *queryEmployeeConfig) {
		config.start = length * page
		config.limit = length
	}
}
func EmployeeFilter(filter pg.Where) EmployeeOption {
	return func(config *queryEmployeeConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func EmployeeOnRow(cb func(*Employee)) EmployeeOption {
	return func(s *queryEmployeeConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Employee(ctx context.Context, opts ...EmployeeOption) (map[interface{}]*Employee, error) {
	config := &queryEmployeeConfig{
		Store:  s,
		filter: pg.NONE(),
		limit:  1000,
		rows:   make(map[interface{}]*Employee),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.selectEmployee(ctx, config, func(row *Employee) {
		config.rows[row.Id] = row
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
func (s *TestStore) selectEmployee(ctx context.Context, config *queryEmployeeConfig, withRow func(*Employee)) error {
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
	SELECT "employee_id", "employee_firstname", "employee_lastname" 
	FROM "tbl_employee"
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
		row := new(Employee)
		err := cursor.Scan(&row.Id, &row.Firstname, &row.Lastname)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
