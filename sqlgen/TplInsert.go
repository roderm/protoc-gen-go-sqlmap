package sqlgen

import "text/template"

var insertTpl = `
func (m *{{ MessageName .  }}) Insert(s *{{ Store }}, ctx context.Context) (error) {
	ins := pg.NewInsert()
	ins.Add(m.{{ GetInsertFieldNames .  ", m." }})

	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
		INSERT INTO "{{ TableName . }}" ( "{{ GetInsertColNames .  "\", \"" }}" )
		VALUES ` + "`" + ` + ins.String() + ` + "`" + `
		RETURNING "{{ getColumnNames . "\", \"" }}"
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
		err := cursor.Scan( &m.{{ getFieldNames . ", &m." }} )
		if err != nil {
			return err
		}
	}
	return nil
}
`

func LoadInsertTemplate(p Printer) *template.Template {
	tpl, err := template.New("Selects").Funcs(GetTemplateFuns(p)).Parse(insertTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}

func (m *Table) Inserter(g Printer) {
	err := LoadInsertTemplate(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
