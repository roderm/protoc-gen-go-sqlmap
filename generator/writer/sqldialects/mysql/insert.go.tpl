{{ if .Config.Create }}
func (m *{{ .MsgName  }}) Insert(s *{{ .StoreName }}, ctx context.Context) (error) {
	query := squirrel.Insert("{{ .Name }}")
	query.SetMap(map[string]interface{}{
		{{- range $i, $f := GetInsertFields . }}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end}}
	})
	_, err := query.RunWith(s.conn).ExecContext(ctx)
	return err
}
{{end}}
