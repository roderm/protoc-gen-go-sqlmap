{{ if .Config.Create }}
func (m *{{ .MsgName  }}) Insert(s *{{ .StoreName }}, ctx context.Context) (error) {
	query = squirrel.Insert("{{ .Name }}")
	query.SetMap(map[string]interface{}{
		{{- range $i, $f := GetInsertFields . }}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end}}
	})
	query.Suffix(`RETURNING {{ getColumnNames . ", " }}`)
	cursor, err := query.RunWith(s.conn).QueryContext(ctx)
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
