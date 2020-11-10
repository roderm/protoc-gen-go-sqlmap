package sqlgen

import (
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
		config.rows[row.Id] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}
	{{ range $i, $f := SubQueries . }}
	if config.load{{ getFieldName $f  }} {
	 	_, err = s.{{ MessageName $f.FK.Remote.Table  }}(ctx, config.opts{{ getFieldName $f }}...)
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
	tpl, err := template.New("Selects").Funcs(TplFuncs).Parse(selectTpl)
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
