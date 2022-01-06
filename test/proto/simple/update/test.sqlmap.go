package update

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

func (m *Employee) Update(s *TestStore, ctx context.Context, conf *pg.UpdateSQL) error {
	base := 1
	if conf == nil {
		conf = &pg.UpdateSQL{
			ValueMap: make(map[string]interface{}),
		}
		conf.ValueMap["employee_firstname"] = m.Firstname
		conf.ValueMap["employee_lastname"] = m.Lastname
	}
	stmt, err := s.conn.PrepareContext(ctx, `
	UPDATE tbl_employee 
	SET `+conf.String(&base)+`
	WHERE "employee_id" = $1
	RETURNING employee_id, employee_firstname, employee_lastname
		`)
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, append([]interface{}{m.GetId()}, conf.Values()...)...)
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
