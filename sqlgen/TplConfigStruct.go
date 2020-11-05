package sqlgen

import (
	"strings"
	"text/template"
)

var configStructTpl = `
type query{{ MessageName . }}Config struct {
	filter pg.Where 
	cb []func(*{{ MessageName . }})
	rows map[string]*{{ MessageName .  }}
	{{ range $index, $sub := SubQueries . }}
	load{{ MessageName $sub }} bool
	opts{{ MessageName $sub }} []{{ MessageName $sub }}Option{{end}}
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
{{range $index, $sub := SubQueries .}}
func {{ MessageName $ }}With{{ MessageName $sub }}(opts ...{{ MessageName $sub }}Option) {{ MessageName $ }}Option {
	return func(config *query{{ MessageName $ }}Config) {
		// {{ TableName $sub }}Tmp := make(map[interface{}]*[]*{{ MessageName $sub }} )
		ids := []interface{}{}
		config.load{{ MessageName $sub }} = true
		config.opts{{ MessageName $sub }} = opts
		config.cb = append(config.cb, func(row *{{ MessageName $ }}) {
			//  {{ TableName $sub }}Tmp[row.{{ GetRemoteFieldname $ $sub }}] = &row.{{ GetFieldnameLinked $ $sub }}
			 ids = append(ids, row.{{ GetRemoteFieldname $ $sub }})
		})
		config.opts{{ MessageName $sub }} = append(config.opts{{ MessageName $sub }}, 
			{{ MessageName $sub }}OnRow(func(row *{{ MessageName $sub }}) {
				{{ if IsReverseFK $ $sub }}
				row.{{ GetDataFieldname $ $sub false }} = config.rows[row.{{ GetDataFieldname $ $sub true }}]
				{{end}}
				config.rows[row.{{ GetDataFieldname $ $sub true }}].{{ GetFieldnameLinked $ $sub }} = append(config.rows[row.{{ GetDataFieldname $ $sub true }}].{{ GetFieldnameLinked $ $sub }}, row)
				// *{{ TableName $sub }}Tmp[row.{{ GetDataFieldname $ $sub true }}] = append(*{{ TableName $sub }}Tmp[row.{{ GetDataFieldname $ $sub true }}], row)
			}),
			{{ MessageName $sub }}Filter(pg.IN("{{ GetDataColname $ $sub}}", ids))) 
	}
}{{ end }}
	
`

func LoadConfigStructTemplate() *template.Template {
	tpl, err := template.New("ConfigStructs").Funcs(template.FuncMap{
		"GetPKName": func(t *Table) string {
			pk := GetPK(t)
			if pk == nil {
				return "someID"
			}
			return pk.desc.GetName()
		},
		"GetPKType": func(t *Table) string {
			pk := GetPK(t)
			if pk == nil {
				return "maybeInt"
			}
			switch pk.desc.GetType().String() {
			case "TYPE_STRING":
				return "string"
			case "TYPE_INT64":
				return "int64"
			case "TYPE_UINT64":
				return "uint64"
			case "TYPE_INT32":
				return "int32"
			case "TYPE_UINT32":
				return "uint32"

			}
			return pk.desc.GetType().String()
		},
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
		"IsReverseFK": func(remote *Table, data *Table) bool {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					if f.desc.IsMessage() {
						mt := strings.Split(f.desc.GetTypeName(), ".")
						return mt[len(mt)-1] == remote.desc.GetName()
					}
				}
			}
			return false
		},
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
		"GetRemoteFieldname": GetRemoteFieldname,
		"GetDataFieldname": func(remote *Table, data *Table, path bool) string {
			for _, f := range data.Cols {
				if remote.Name == f.dbfkTable {
					if f.desc.IsMessage() && path {
						return f.desc.GetName() + "." + GetRemoteFieldname(remote, data)
					} else {
						return f.desc.GetName()
					}
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

func GetPK(t *Table) *field {
	for _, c := range t.Cols {
		if c.PK != PK_NONE {
			return c
		}
	}
	return nil

}
func GetRemoteFieldname(remote *Table, data *Table) string {
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
}
