{{ if .Config.Delete }}
func (m *{{ .MsgName  }}) Delete(s *{{ .StoreName }}, ctx context.Context) (error) {
	query := squirrel.Delete("{{ .Name }}").Where(squirrel.Eq{
		{{- range $i, $f := GetPrimaries .}}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end }}
	})
	
	_, err := query.RunWith(s.conn).ExecContext(ctx)
	return err
}
{{end}}
