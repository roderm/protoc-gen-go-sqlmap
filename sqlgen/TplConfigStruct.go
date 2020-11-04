package sqlgen

import (
	"strings"
	"text/template"
)

var configStructTpl = `
type query{{ MessageName . }}Config struct {
	filter pg.Where 
	cb []func(*{{ MessageName . }})
	rows []*{{ MessageName .  }}
	{{ range $index, $sub := SubQueries . }}
	load{{ MessageName $sub }} bool
	opts{{ MessageName $sub }} []{{ MessageName $sub }}Option{{end}}
}
type {{ MessageName . }}Option func(*query{{ MessageName . }}Config)
func {{ MessageName . }}Filter(filter pg.Where) {{ MessageName . }}Option {
	func(config *query{{ MessageName . }}Config) {
		if config.filter == nil {
			config.filter = filter
		} else {
			pg.AND(config.filter, filter)
		}
	}
}

func {{ MessageName . }}OnRow(cb func(*{{ MessageName . }})) Select{{ MessageName . }}Option {
	return func(s *select{{ MessageName . }}Config){
		s.cb = append(s.cb, cb)
	}
}
{{range $index, $sub := SubQueries .}}
func {{ MessageName $ }}With{{ MessageName $sub }}(opts ...{{ MessageName $sub }}Option) {{ MessageName $ }}Option {
	// TODO: Need to check if this works or find other solution
	{{ TableName $sub }}Tmp := make(map[interface{}]*[]*{{ MessageName $sub }} )
	ids := []interface{}{}
	func(config *query{{ MessageName $ }}Config) {
		config.load{{ MessageName $sub }} = true
		config.opts{{ MessageName $sub }} = opts
		config.cb = append(config.cb, func(row *{{ MessageName $ }}) {
			 {{ TableName $sub }}Tmp[row.{{ GetRemoteFieldname $ $sub }}] = &row.{{ GetFieldnameLinked $ $sub }}
			 ids = append(ids, row.{{ GetRemoteFieldname $ $sub }})
		})
		config.opts{{ MessageName $sub }} = append(config.opts{{ MessageName $sub }}, 
			{{ MessageName $sub }}OnRow(func(row *{{ MessageName $sub }}) {
				*{{ TableName $sub }}Tmp[row.{{ GetRemoteFieldname $ $sub }}] = append(*{{ TableName $sub }}Tmp[row.{{ GetRemoteFieldname $ $sub }}], row)
			}),
			{{ MessageName $sub }}Filter(pg.IN("{{ GetDataColname $ $sub}}", ids))) 
	}
}{{ end }}
	
`

func LoadConfigStructTemplate() *template.Template {
	tpl, err := template.New("ConfigStructs").Funcs(template.FuncMap{
		"GetFieldnameLinked": func(remote *Table, data *Table) string {
			for _, rf := range remote.Cols {
				mt := strings.Split(rf.desc.GetTypeName(), ".")
				if mt[len(mt)-1] == data.desc.GetName() {
					return rf.desc.GetName()
				}
			}
			return "notFound"
		},
		// "GetTargetFieldname": func(remote *Table, data *Table) string {
		// 	for _, f := range data.Cols {
		// 		if remote.Name == f.dbfkTable {
		// 			for _, rf := range remote.Cols {
		// 			}
		// 		}
		// 	}
		// },
		"GetRemoteColname": func(remote *Table, data *Table) string {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					for _, rf := range remote.Cols {
						if rf.ColName == f.dbfkField {
							return rf.ColName
						}
					}
				}
			}
			return ""
		},
		"GetDataColname": func(remote *Table, data *Table) string {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					return f.ColName
				}
			}
			return ""
		},
		"GetRemoteFieldname": func(remote *Table, data *Table) string {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					for _, rf := range remote.Cols {
						if rf.ColName == f.dbfkField {
							return rf.desc.GetName()
						}
					}
				}
			}
			return ""
		},
		"GetDataFieldname": func(remote *Table, data *Table) string {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					return f.desc.GetName()
				}
			}
			return ""
		},
		"MessageName": func(t *Table) string {
			return t.desc.GetName()
		},
		"TableName": func(t *Table) string {
			return t.Name
		},
		"SubQueries": func(t *Table) []*Table {
			tables := []*Table{}
			for _, f := range t.Cols {
				for _, fk := range f.FK {
					tables = append(tables, fk.Target.Table)
				}
			}
			return tables
		},
		"getFKMessages": func(t *Table) map[*field]fieldFK {
			res := make(map[*field]fieldFK)
			for _, f := range t.Cols {
				if f.desc.IsMessage() || f.desc.IsRepeated() {
					fk, err := TableMessageStore.GetFKfromType(f)
					if err == nil {
						res[f] = fk
					}
				}
			}
			return res
		},
		"getColumnNames": func(t *Table, separator string) string {
			str := ""
			for _, f := range t.GetOrderedCols() {
				if len(f.FK) == 0 && len(f.ColName) > 0 {
					str = str + f.ColName + separator
				}
			}
			return strings.TrimSuffix(str, separator)
		},
		"getFieldNames": func(t *Table, separator string) string {
			str := ""
			for _, f := range t.GetOrderedCols() {
				if len(f.FK) == 0 && len(f.ColName) > 0 {
					str = str + f.desc.GetName() + separator
				}
			}
			return strings.TrimSuffix(str, separator)
		},
		"getFieldName": func(f *field) string {
			return f.desc.GetName()
		},
		"getColumnName": func(f *field) string {
			return f.ColName
		},
	}).Parse(configStructTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}
func (m *Table) ConfigStructs(g Printer) {
	err := LoadConfigStructTemplate().Execute(g, m)
	if err != nil {
		panic(err)
	}
}
