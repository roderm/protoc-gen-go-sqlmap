package delete

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

func (m *Employee) Delete(s *TestStore, ctx context.Context) error {
	query := squirrel.Delete("tbl_employee").Where(squirrel.Eq{
		"employee_id": m.Id,
	})
	query.Suffix(`RETURNING "employee_id", "employee_firstname", "employee_lastname"`)

	cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close()
	resultRows := []*Employee{}
	err = sqlx.StructScan(cursor, &resultRows)
	if err != nil {
		return fmt.Errorf("sqlx.StructScan failed: %s", err)
	}
	if len(resultRows) > 0 {
		m = resultRows[0]
	} else {
		err = fmt.Errorf("can't get deleted col")
	}
	return err
}
