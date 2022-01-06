{{ if .Config.Read }}
type query{{ .MsgName }}Config struct {
	Store *{{ .StoreName }}
	filter pg.Where 
	start int
	limit int
	beforeReturn []func(map[interface{}]*{{ .MsgName  }}) error
	cb []func(*{{ .MsgName }})
	rows map[interface{}]*{{ .MsgName  }}
	{{- range $i, $join := $.Joins }}
	load{{ $join.TargetFieldName }} bool
	opts{{ $join.TargetFieldName }} []{{ $join.SourcePackagePrefix}}{{ $join.SourceMessageName }}Option
	{{- end }}
}

type {{ .MsgName }}Option func(*query{{ .MsgName }}Config)
func {{ .MsgName }}Paging(page, length int) {{ .MsgName }}Option {
	return func(config *query{{ .MsgName }}Config) {
		config.start = length * page
		config.limit = length
	}
}
func {{ .MsgName }}Filter(filter pg.Where) {{ .MsgName }}Option {
	return func(config *query{{ .MsgName }}Config) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
		}
	}
}

func {{ .MsgName }}OnRow(cb func(*{{ .MsgName }})) {{ .MsgName }}Option {
	return func(s *query{{ .MsgName }}Config){
		s.cb = append(s.cb, cb)
	}
}
{{end}}
