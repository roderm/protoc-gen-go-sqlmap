package sqlgen

import "text/template"

var insertTpl = `
func (m *{{ MessageName .  }}) Insert(s *Store, ctx context.Context) (error) {
	ins := pg.NewInsert()
	ins.Add(v.{{ GetInsertFieldNames .  ", v." }})

	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
		INSERT INTO {{ TableName . }} ( {{ GetInsertColNames .  ", " }} )
		VALUES ` + "`" + ` + ins.String() + ` + "`" + `
		RETURNING {{ getColumnNames . ", " }}
		` + "`" + `)
	
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, ins.Values()...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		err := cursor.Scan( m.{{ getFieldNames . ", m." }} )
		if err != nil {
			return err
		}
	}
	return nil
}
`

// func (s *Store) Insert(ctx context.Context, values ...*{{ MessageName .  }}) (error) {
// 	ins := pg.NewInsert()
// 	for _, v := range values {
// 		ins.Add(v.{{ GetInsertFieldNames .  ", v." }})
// 	}

// 	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
// 		INSERT INTO {{ TableName . }} ( {{ GetInsertColNames .  ", " }} )
// 		VALUES ` + "`" + ` + ins.String() + ` + "`" + `
// 		RETURNING {{ getColumnNames . ", " }}
// 		` + "`" + `)

// 	if err != nil {
// 		return err
// 	}

// 	cursor, err := stmt.QueryContext(ctx, ins.Values()...)
// 	if err != nil {
// 		return err
// 	}
// 	defer cursor.Close()
// 	return nil
// }

/*
func (s *Store) {{ MessageName .  }}Insert(ctx context.Context, values ...*{{ MessageName .  }}) (error) {
	ins := pg.NewInsert()
	for _, v := range values {
		ins.Add(v.{{ GetInsertFieldNames .  ", v." }})
	}
	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	INSERT INTO {{ TableName . }} ( {{ GetInsertColNames .  ", " }} )
	VALUES ` + "`" + ` + ins.String() + ` + "`" + `
	RETURNING {{ getColumnNames . ", " }}
	` + "`" + `)

	cursor, err := stmt.QueryContext(ctx, ins.Values()...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for cursor.Next() {
		var row {{ MessageName .  }}
		err := cursor.Scan( &row.{{ getFieldNames . ", &row." }} )
		if err != nil {
			return err
		}
		func(update *{{ MessageName .  }}) {
			for _, old := range values {
				if old.id == update.id {
					{{ range GetColumns . }}
						old.{{ . }} = update.{{ . }}{{ end }}
					{{range $name, $field := getFKMessages . }}
						for _, child := range old.{{ getFieldName $name }} {
							child.{{ getFieldName $field.Target }} = update.{{ getFieldName $field.Source }}
						}{{ end }}
				}
			}
		} (&row)
	}
	return nil
}
*/

func LoadInsertTemplate() *template.Template {
	tpl, err := template.New("Selects").Funcs(TplFuncs).Parse(insertTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}

func (m *Table) Inserter(g Printer) {
	err := LoadInsertTemplate().Execute(g, m)
	if err != nil {
		panic(err)
	}
}
