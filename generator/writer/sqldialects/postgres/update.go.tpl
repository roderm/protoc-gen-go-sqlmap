{{ if .Config.Update }}
func (m *{{ .MsgName  }}) Update(s *{{ .StoreName }}, ctx context.Context) (error) {
	query := squirrel.Update("{{ .Name }}").Where(squirrel.Eq{
		{{- range $i, $f := GetPrimaries .}}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end }}
	})
	{{- range $i, $f := GetUpdateFields . }}
	query.Set("{{ $f.ColName }}", m.{{ $f.MsgName }})
	{{- end}}
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
		err = fmt.Errorf("can't get updated col") 
	}
	return err
}
{{ end }}
