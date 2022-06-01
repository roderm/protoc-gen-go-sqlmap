{{ if .Config.Update }}
func (m *{{ .MsgName  }}) Update(s *{{ .StoreName }}, ctx context.Context) (error) {
	filter := squirrel.Eq{
		{{- range $i, $f := GetPrimaries .}}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end }}
	}
	query := squirrel.Update("{{ .Name }}").Where(filter)
	{{- range $i, $f := GetUpdateFields . }}
	query.Set("{{ $f.ColName }}", m.{{ $f.MsgName }})
	{{- end}}
	
	_, err := query.RunWith(s.conn).ExecContext(ctx)
	return err
}
{{ end }}
