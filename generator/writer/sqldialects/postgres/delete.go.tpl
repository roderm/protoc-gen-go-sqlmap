{{ if .Config.Delete }}
func (m *{{ .MsgName  }}) Delete(s *{{ .StoreName }}, ctx context.Context) (error) {
	query := squirrel.Delete("{{ .Name }}").Where(squirrel.Eq{
		{{- range $i, $f := GetPrimaries .}}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end }}
	})
	query.Suffix(`RETURNING {{ getColumnNames . ", " }}`)

	cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer cursor.Close()
	resultRows := []*{{ .MsgName  }}{}
	err = sqlx.StructScan(cursor, &resultRows)
	if err != nil {
		return fmt.Errorf("sqlx.StructScan failed: %s", err)
	}
	if len(resultRows) > 0 {
		m = resultRows[0]
	} else {
		err = fmt.Errorf("can't get deleted col") 
	}
	return err
}
{{end}}
