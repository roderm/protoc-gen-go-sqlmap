package sqlgen

import "text/template"

var updateTpl = `
func (m *{{ MessageName .  }}) Update(s *{{ Store }}, ctx context.Context) (error) {

	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	UPDATE {{ TableName . }} 
	WHERE {{ GetPKCol . }} = $1
	RETURNING {{ getColumnNames . ", " }}
		` + "`" + `)
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, m.{{ GetPKName . }})
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

func LoadUpdateTemplate(p Printer) *template.Template {
	tpl, err := template.New("Update").Funcs(GetTemplateFuns(p)).Parse(updateTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}

func (m *Table) Updater(g Printer) {
	err := LoadUpdateTemplate(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
