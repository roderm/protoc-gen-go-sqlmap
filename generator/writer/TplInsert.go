package writer

import (
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
)

var insertTpl = `
{{ if .Create }}
func (m *{{ MessageName .  }}) Insert(s *{{ Store }}, ctx context.Context) (error) {
	ins := pg.NewInsert()
	ins.Add(m.{{ GetInsertFieldNames .  ", m." }})

	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
		INSERT INTO "{{ TableName . }}" ( "{{ GetInsertColNames .  "\", \"" }}" )
		VALUES ` + "`" + ` + ins.String(nil) + ` + "`" + `
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
{{end}}
`

func LoadInsertTemplate(p Printer) *template.Template {
	tpl, err := template.New("Selects").Funcs(GetTemplateFuns(p)).Parse(insertTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}

func WriteInsertes(g Printer, m *types.Table) {
	err := LoadInsertTemplate(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
