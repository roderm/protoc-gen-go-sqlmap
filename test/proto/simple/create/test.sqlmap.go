package create

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

type EmployeeList map[interface{}]*Employee

func (m *Employee) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		"employee_id": m.Id,
	}
	return pk
}
func (m *Employee) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => 'Employee' failed: %s", string(buff), err)
	}
	return nil
}

func (m *Employee) Insert(s *TestStore, ctx context.Context) error {
	cursor, err := s.conn.NamedQueryContext(ctx, `
		INSERT INTO "tbl_employee" ( "employee_firstname", "employee_lastname" )
		VALUES ( :employee_firstname, :employee_lastname )
		RETURNING "employee_id", "employee_firstname", "employee_lastname";`,
		m,
	)
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
