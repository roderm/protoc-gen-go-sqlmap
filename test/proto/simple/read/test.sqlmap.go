package read

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

type queryEmployeeConfig struct {
	Store        *TestStore
	filter       []squirrel.Sqlizer
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

func EmployeeFilter(filter ...squirrel.Sqlizer) EmployeeOption {
	return func(config *queryEmployeeConfig) {
		config.filter = append(config.filter, filter...)
	}
}

func EmployeeOnRow(cb func(*Employee)) EmployeeOption {
	return func(s *queryEmployeeConfig) {
		s.cb = append(s.cb, cb)
	}
}

func (s *TestStore) Employee(ctx context.Context, opts ...EmployeeOption) (EmployeeList, error) {
	config := &queryEmployeeConfig{
		Store:  s,
		filter: squirrel.And{},
		limit:  1000,
		rows:   make(EmployeeList),
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

func (s *TestStore) GetEmployeeSelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`"employee_id", "employee_firstname", "employee_lastname"`).
		From("\"tbl_employee\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *TestStore) selectEmployee(ctx context.Context, config *queryEmployeeConfig, withRow func(*Employee)) error {
	query := s.GetEmployeeSelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'selectEmployee': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*Employee{}
	for cursor.Next() {
		row := new(Employee)
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
