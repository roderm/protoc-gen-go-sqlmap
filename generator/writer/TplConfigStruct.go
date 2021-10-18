package writer

import (
	"text/template"
)

var configStructTpl = `
{{if .JSONB }}
func (m *{{ MessageName . }}) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	return json.Unmarshal(buff, m)
}

func (m *{{ MessageName . }}) Value() (driver.Value, error) {
	return json.Marshal(m)
}
{{else}}
func (m *{{ MessageName . }}) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.{{ GetPKName . }} = {{ GetPKType . }}(buff)
	return nil
}

func (m *{{ MessageName . }}) Value() (driver.Value, error) {
	return m.{{ GetPKName . }}, nil
}
{{end}}


{{ if .Read }}
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

type {{ MessageName . }}Option func(*query{{ MessageName . }}Config)
func {{ MessageName . }}Filter(filter pg.Where) {{ MessageName . }}Option {
	return func(config *query{{ MessageName . }}Config) {
		if config.filter == nil {
			config.filter = filter
		} else {
			config.filter = pg.AND(config.filter, filter)
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
		map{{ getFieldName $f }} := make(map[string]*{{ MessageName $ }})
		config.load{{ getFieldName $f }} = true
		config.opts{{ getFieldName $f }} = opts
		config.cb = append(config.cb, func(row *{{ MessageName $ }}) {
			{{if IsRepeated $f }}
				map{{ getFieldName $f }}[row.{{ GetPKName $ }}] = row
			{{else}} 
				map{{ getFieldName $f }}[row.{{ getFullFieldName $f }}] = row
			{{end}}
		})
		config.opts{{ getFieldName $f }} = append(config.opts{{ getFieldName $f }}, 
			{{ MessageName $f.FK.Remote.Table }}OnRow(func(row *{{ MessageName $f.FK.Remote.Table }}) {
				{{ if IsReverseFK $f }}
				if config.rows[row.{{ getFieldName $f }}] != nil {
					row.{{ getFieldName $f }} = config.rows[row.{{ getFieldName $f }}]
				}
				{{end}}

				{{if IsRepeated $f }}
				if config.rows[row.{{ getFullFieldName $f.FK.Remote }}] != nil {
					config.rows[row.{{ getFullFieldName $f.FK.Remote }}].{{ getFieldName $f }} = append(config.rows[row.{{ getFullFieldName $f.FK.Remote }}].{{ getFieldName $f }}, row)
				}
				{{else}}
				item := map{{ getFieldName $f }}[row.{{ GetPKName $f.FK.Remote.Table }}]
				if config.rows[item.{{ GetPKName $ }}] != nil {
					config.rows[item.{{ GetPKName $ }}].{{ getFieldName $f }} = row
				}
				{{end}}
			}),
			{{ MessageName $f.FK.Remote.Table }}Filter(pg.INCallabel("{{ $f.DbfkField }}", func() []interface{} {
				ids := []interface{}{}
				for id := range map{{ getFieldName $f }} {
					ids = append(ids, id)
				}
				return ids
			})),
		) 
	}
}{{ end }}
{{end}}
`

func LoadConfigStructTemplate(p Printer) *template.Template {
	tpl, err := template.New("ConfigStructs").Funcs(GetTemplateFuns(p)).Parse(configStructTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}
