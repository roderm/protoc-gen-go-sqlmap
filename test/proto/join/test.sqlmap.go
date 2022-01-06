package join

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

type queryEmployeeConfig struct {
	Store        *TestStore
	filter       pg.Where
	start        int
	limit        int
	beforeReturn []func(map[interface{}]*Employee) error
	cb           []func(*Employee)
	rows         map[interface{}]*Employee
	loadManager  bool
	optsManager  []EmployeeOption
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

func EmployeeWithManager(opts ...EmployeeOption) EmployeeOption {
	return func(config *queryEmployeeConfig) {
		config.loadManager = true
		parent := make(map[interface{}][]*Employee)
		config.cb = append(config.cb, func(row *Employee) {
			child_key := row.GetManager().EmployeeID
			parent[child_key] = append(parent[child_key], row)
		})
		config.optsManager = append(opts,
			EmployeeFilter(
				pg.INCallabel("employee_id", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			EmployeeOnRow(func(row *Employee) {
				children := parent[row.EmployeeID]
				for _, c := range children {
					c.Manager = row
				}
			}),
		)
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
		config.rows[row.EmployeeID] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}

	if config.loadManager {
		// github.com/roderm/protoc-gen-go-sqlmap/test/proto/join

		_, err = s.Employee(ctx, config.optsManager...)

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

func (s *TestStore) GetEmployeeSelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "employee_id", "employee_firstname", "employee_lastname", "employee_manager"
		FROM "tbl_employee"
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

func (s *TestStore) selectEmployee(ctx context.Context, config *queryEmployeeConfig, withRow func(*Employee)) error {
	query, vals := s.GetEmployeeSelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'selectEmployee': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'selectEmployee' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &Employee{
			Manager: new(Employee),
		}
		err := cursor.Scan(&row.EmployeeID, &row.Firstname, &row.Lastname, &row.GetManager().EmployeeID)
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
