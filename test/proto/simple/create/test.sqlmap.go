// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: test/proto/simple/create/test.proto

package create

import (
	context "context"
	sql "database/sql"
	driver "database/sql/driver"
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

func (m *Employee) Insert(s *TestStore, ctx context.Context) error {
	ins := pg.NewInsert()
	ins.Add(m.Firstname, m.Lastname)

	stmt, err := s.conn.PrepareContext(ctx, `
		INSERT INTO "tbl_employee" ( "employee_firstname", "employee_lastname" )
		VALUES `+ins.String()+`
		RETURNING "employee_id", "employee_firstname", "employee_lastname"
		`)

	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, ins.Values()...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		err := cursor.Scan(&m.Id, &m.Firstname, &m.Lastname)
		if err != nil {
			return err
		}
	}
	return nil
}
