// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: test/proto/simple/read/test.proto

package read

import (
	context "context"
	sql "database/sql"
	driver "database/sql/driver"
	json "encoding/json"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	pg "github.com/roderm/protoc-gen-go-sqlmap/lib/pg"
	_ "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

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
		return fmt.Errorf("Failed % ", value)
	}
	m.Id = string(buff)
	return nil
}

func (m *Employee) Value() (driver.Value, error) {
	return m.Id, nil
}

type queryEmployeeConfig struct {
	Store        *TestStore
	filter       pg.Where
	beforeReturn []func(map[string]*Employee) error
	cb           []func(*Employee)
	rows         map[string]*Employee
}

type EmployeeOption func(*queryEmployeeConfig)

func EmployeeFilter(filter pg.Where) EmployeeOption {
	return func(config *queryEmployeeConfig) {
		if config.filter == nil {
			config.filter = filter
		} else {
			pg.AND(config.filter, filter)
		}
	}
}

func EmployeeOnRow(cb func(*Employee)) EmployeeOption {
	return func(s *queryEmployeeConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Employee(ctx context.Context, opts ...EmployeeOption) (map[string]*Employee, error) {
	config := &queryEmployeeConfig{
		Store:  s,
		filter: pg.NONE(),
		rows:   make(map[string]*Employee),
	}
	for _, o := range opts {
		o(config)
	}

	err := s.selectEmployee(ctx, config.filter, func(row *Employee) {
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
func (s *TestStore) selectEmployee(ctx context.Context, filter pg.Where, withRow func(*Employee)) error {
	where, vals := pg.GetWhereClause(filter)
	stmt, err := s.conn.PrepareContext(ctx, `
	SELECT "employee_id", "employee_firstname", "employee_lastname" 
	FROM "tbl_employee"
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
		row := new(Employee)
		err := cursor.Scan(&row.Id, &row.Firstname, &row.Lastname)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
