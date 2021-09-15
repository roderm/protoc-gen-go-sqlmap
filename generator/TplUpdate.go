package generator

import (
	"text/template"
)

var updateTpl = `
{{ if .Update }}
func (m *{{ MessageName .  }}) Update(s *{{ Store }}, ctx context.Context, conf *pg.UpdateSQL) (error) {
	base := 1
	if conf == nil {
		conf = &pg.UpdateSQL{
			ValueMap: make(map[string]interface{}),
		}{{ range $i, $f := getInsertFields .}}
		conf.ValueMap["{{getColumnName $f}}"] = m.{{getFieldName $f}}{{end}}
	}
	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	UPDATE {{ TableName . }} 
	SET ` + "`" + ` + conf.String(&base) + ` + "`" + `
	WHERE "{{ GetPKCol . }}" = $1
	RETURNING {{ getColumnNames . ", " }}
		` + "`" + `)
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, append([]interface{}{ m.{{ GetPKName . }} }, conf.Values()...)... )
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
{{ end }}
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
