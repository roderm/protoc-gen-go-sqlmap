package writer

import (
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
)

var deleteTpl = `
{{ if .Delete }}
func (m *{{ MessageName .  }}) Delete(s *{{ Store }}, ctx context.Context) (error) {

	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	DELETE FROM "{{ TableName . }}"
	WHERE "{{ GetPKCol . }}" = $1
	RETURNING "{{ getColumnNames . "\", \"" }}"
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
{{end}}
`

func LoadDeleteTemplate(p Printer) *template.Template {
	tpl, err := template.New("Delete").Funcs(GetTemplateFuns(p)).Parse(deleteTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}

func WriteDeletes(g Printer, m *types.Table) {
	err := LoadDeleteTemplate(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
