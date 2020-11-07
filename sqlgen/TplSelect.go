package sqlgen

import (
	"strings"
	"text/template"
)

var selectTpl = `
func (s *Store) {{ MessageName .  }}(ctx context.Context, opts ...{{ MessageName .  }}Option) (map[string]*{{ MessageName .  }}, error) {
	config := &query{{ MessageName .  }}Config{
		filter: pg.NONE(),
		rows: make(map[string]*{{ MessageName .  }}),
	}
	for _, o := range opts {
		o(config)
	}

	err := s.select{{ MessageName .  }}(ctx, config.filter, func(row *{{ MessageName .  }}) {
		for _, cb := range config.cb {
			cb(row)
		}
		config.rows[row.Id] = row
	})
	if err != nil {
		return config.rows, err
	}
	{{ range $index, $sub := SubQueries . }}
	if config.load{{ MessageName .  }} {
	 	_, err = s.{{ MessageName .  }}(ctx, config.opts{{ GetRemoteFieldname $ $sub }}...)
	}
	if err != nil {
	 	return config.rows, err
	}
	{{ end }}
	return config.rows, nil
}
func (s *Store) select{{ MessageName . }}(ctx context.Context, filter pg.Where, withRow func(*{{ MessageName .  }})) error {
	where, vals := pg.GetWhereClause(filter)
	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	SELECT {{ getColumnNames .  ", " }} 
	FROM {{ TableName . }}
	` + "`" + `+where)
	if err != nil {
		return err
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		row := new({{ MessageName .  }})
		err := cursor.Scan( &row.{{ getFieldNames . ", &row." }} )
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
`

// func (s *Store) {{ MessageName .  }}(ctx context.Context, filter pg.Where{{range $name, $field := getFKMessages . }}, employeeFilter pg.Where {{ end }}) ([]*{{ MessageName .  }}, error) {
// 	rows := []*{{ MessageName .  }}{}
// 	queries_id := []interface{}{} // must be generated
// 	{{range $name, $field := getFKMessages . }}
// 	{{ TableName $field.Source.Table }}Tmp := make(map[interface{}]*[]*{{ MessageName $field.Source.Table }} ) {{end}}
//
// 	err := s.select{{ MessageName .  }}(ctx, filter, func(row *{{ MessageName .  }}) {
// 		rows = append(rows, row)
// 		queries_id = append(queries_id, row.Id) // must be generated
//
// 		{{range $name, $field := getFKMessages . }}
// 			{{ TableName $field.Source.Table }}Tmp[ row.{{ getFieldName $field.Target }}] = &row.{{ getFieldName $name }} {{end}}
// 	})
//
// 	{{range $name, $field := getFKMessages . }}
// 	err = s.select{{ MessageName $field.Source.Table }}(ctx, pg.AND(pg.IN("{{ getColumnName $field.Source }}", queries_id...), {{ TableName $field.Source.Table }}Filter), func(row *{{ MessageName $field.Source.Table }}) {
// 		*{{ TableName $field.Source.Table }}Tmp[row.{{ getFieldName $field.Source }}] = append(*{{ TableName $field.Source.Table }}Tmp[row.{{ getFieldName $field.Source }}], row)
// 	})
// 	if err != nil {
// 		return rows, err
// 	}
// 	{{ end }}
// 	return rows, nil
// }

func LoadSelectTemplate() *template.Template {
	tpl, err := template.New("Selects").Funcs(template.FuncMap{
		"SubQueries": func(t *Table) []*Table {
			tables := []*Table{}
			for _, f := range t.Cols {
				for _, fk := range f.DepFKs {
					tables = append(tables, fk.Target.Table)
				}
			}
			return tables
		},
		"GetRemoteFieldname": GetRemoteFieldname,
		"GetDataFieldname": func(remote *Table, data *Table, path bool) string {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					if f.desc.IsMessage() && path {
						return f.desc.GetName() + "." + GetRemoteFieldname(remote, data)
					} else {
						return f.desc.GetName()
					}
				}
			}
			return ""
		},
		"MessageName": func(t *Table) string {
			return t.desc.GetName()
		},
		"TableName": func(t *Table) string {
			return t.Name
		},
		"getFKMessages": func(t *Table) map[*field]*fieldFK {
			res := make(map[*field]*fieldFK)
			for _, f := range t.Cols {
				if f.desc.IsMessage() || f.desc.IsRepeated() {
					fk, err := TableMessageStore.GetFKfromType(f)
					if err == nil {
						res[f] = fk
					}
				}
			}
			return res
		},
		"getColumnNames": func(t *Table, separator string) string {
			str := ""
			for _, f := range t.GetOrderedCols() {
				if len(f.ColName) > 0 {
					str = str + f.ColName + separator
				}
			}
			return strings.TrimSuffix(str, separator)
		},
		"getFieldNames": func(t *Table, separator string) string {
			str := ""
			for _, f := range t.GetOrderedCols() {
				if len(f.ColName) > 0 {
					str = str + f.desc.GetName() + separator
				}
			}
			return strings.TrimSuffix(str, separator)
		},
		"getFieldName": func(f *field) string {
			return f.desc.GetName()
		},
		"getColumnName": func(f *field) string {
			return f.ColName
		},
	}).Parse(selectTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}
func (m *Table) Querier(g Printer) {
	err := LoadSelectTemplate().Execute(g, m)
	if err != nil {
		panic(err)
	}
}
