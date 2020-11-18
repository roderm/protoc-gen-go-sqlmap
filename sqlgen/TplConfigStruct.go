package sqlgen

import (
	"text/template"
)

var configStructTpl = `
type query{{ MessageName . }}Config struct {
	Store *{{ Store }}
	filter pg.Where 
	beforeReturn []func(map[string]*{{ MessageName .  }}) error
	cb []func(*{{ MessageName . }})
	rows map[string]*{{ MessageName .  }}
	{{ range $i, $f := SubQueries . }}
	load{{ getFieldName $f }} bool
	opts{{ getFieldName $f }} []{{ MessageName $f.FK.Remote.Table }}Option{{end}}
}
// TODO: get correct field and type
// Scan for SQL, only works for now on PK "id" and type string
func (m *{{ MessageName . }}) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed % ", value)
	}
	m.{{ GetPKName . }} = {{ GetPKType . }}(buff)
	return nil
}

type {{ MessageName . }}Option func(*query{{ MessageName . }}Config)
func {{ MessageName . }}Filter(filter pg.Where) {{ MessageName . }}Option {
	return func(config *query{{ MessageName . }}Config) {
		if config.filter == nil {
			config.filter = filter
		} else {
			pg.AND(config.filter, filter)
		}
	}
}

func {{ MessageName . }}OnRow(cb func(*{{ MessageName . }})) {{ MessageName . }}Option {
	return func(s *query{{ MessageName . }}Config){
		s.cb = append(s.cb, cb)
	}
}
{{range $i, $f := SubQueries . }}
func {{ MessageName $ }}With{{ getFieldName $f }}(opts ...{{ MessageName $f.FK.Remote.Table }}Option) {{ MessageName $ }}Option {
	return func(config *query{{ MessageName $ }}Config) {
		ids := []interface{}{}
		config.load{{ getFieldName $f }} = true
		config.opts{{ getFieldName $f }} = opts
		config.cb = append(config.cb, func(row *{{ MessageName $ }}) {
			 ids = append(ids, row.Id)
		})
		config.opts{{ getFieldName $f }} = append(config.opts{{ getFieldName $f }}, 
			{{ MessageName $f.FK.Remote.Table }}OnRow(func(row *{{ MessageName $f.FK.Remote.Table }}) {
				{{ if IsReverseFK $f }}
				row.{{ getFieldName $f }} = config.rows[row.{{ getFieldName $f }}]
				{{end}}

				{{if IsRepeated $f }}
				config.rows[row.{{ getFullFieldName $f.FK.Remote }}].{{ getFieldName $f }} = append(config.rows[row.{{ getFullFieldName $f.FK.Remote }}].{{ getFieldName $f }}, row)
				{{end}}
			}),
			{{ MessageName $f.FK.Remote.Table }}Filter(pg.IN("{{ $f.DbfkField }}", ids))) 
	}
}{{ end }}
	
`

func LoadConfigStructTemplate(p Printer) *template.Template {
	tpl, err := template.New("ConfigStructs").Funcs(GetTemplateFuns(p)).Parse(configStructTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}
