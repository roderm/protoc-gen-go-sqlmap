package create

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

func (m *Employee) Insert(s *TestStore, ctx context.Context) error {
	ins := pg.NewInsert()
	ins.Add(m.Firstname, m.Lastname)

	stmt, err := s.conn.PrepareContext(ctx, `
		INSERT INTO "tbl_employee" ( "employee_firstname", "employee_lastname" )
		VALUES `+ins.String(nil)+`
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
