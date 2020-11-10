package sqlgen

var insertTpl = `
func (s *Store) {{ MessageName .  }}Insert(ctx context.Context, values ...*{{ MessageName .  }}) (error) {
	ins := pg.NewInsert()
	for _, v := range values {
		ins.Add(v.{{ GetInsertFieldNames .  ", v." }})
	}
	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	INSERT INTO {{ TableName . }} ( {{ GetInsertColNames .  ", " }} )
	VALUES ` + "`" + ` + ins.String() + ` + "`" + `
	RETURNING {{ getColumnNames . ", " }}
	` + "`" + `)

	cursor, err := stmt.QueryContext(ctx, ins.Values()...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	for cursor.Next() {
		var row {{ MessageName .  }}
		err := cursor.Scan( &row.{{ getFieldNames . ", &row." }} )
		if err != nil {
			return err
		}
		func(update *{{ MessageName .  }}) {
			for _, old := range values {
				if old.id == update.id {
					{{ range GetColumns . }}
						old.{{ . }} = update.{{ . }}{{ end }}
					{{range $name, $field := getFKMessages . }}
						for _, child := range old.{{ getFieldName $name }} {
							child.{{ getFieldName $field.Target }} = update.{{ getFieldName $field.Source }}
						}{{ end }}
				}
			}
		} (&row)
	}
	return nil
}
`

// func LoadInsertTemplate() *template.Template {
// 	tpl, err := template.New("Selects").Funcs(template.FuncMap{
// 		"GetInsertFieldNames": func(t *Table, separator string) string {
// 			str := ""
// 			for _, f := range t.GetOrderedCols() {
// 				if len(f.DepFKs) == 0 && len(f.ColName) > 0 && f.PK == PK_NONE {
// 					str = str + f.desc.GetName() + separator
// 				}
// 			}
// 			return strings.TrimSuffix(str, separator)
// 		},
// 		"GetInsertColNames": func(t *Table, separator string) string {
// 			str := ""
// 			for _, f := range t.GetOrderedCols() {
// 				if len(f.DepFKs) == 0 && len(f.ColName) > 0 && f.PK == PK_NONE {
// 					str = str + f.ColName + separator
// 				}
// 			}
// 			return strings.TrimSuffix(str, separator)
// 		},
// 		"MessageName": func(t *Table) string {
// 			return t.desc.GetName()
// 		},
// 		"TableName": func(t *Table) string {
// 			return t.Name
// 		},
// 		"getFKMessages": func(t *Table) map[*field]*fieldFK {
// 			res := make(map[*field]*fieldFK)
// 			for _, f := range t.Cols {
// 				if f.desc.IsMessage() || f.desc.IsRepeated() {
// 					fk, err := TableMessageStore.GetFKfromType(f)
// 					if err == nil {
// 						res[f] = fk
// 					}
// 				}
// 			}
// 			return res
// 		},
// 		"GetColumns": func(t *Table) []string {
// 			result := []string{}
// 			for _, f := range t.GetOrderedCols() {
// 				if len(f.DepFKs) == 0 && len(f.ColName) > 0 {
// 					result = append(result, f.ColName)
// 				}
// 			}
// 			return result
// 		},
// 		"getColumnNames": func(t *Table, separator string) string {
// 			str := ""
// 			for _, f := range t.GetOrderedCols() {
// 				if len(f.DepFKs) == 0 && len(f.ColName) > 0 {
// 					str = str + f.ColName + separator
// 				}
// 			}
// 			return strings.TrimSuffix(str, separator)
// 		},
// 		"getFieldNames": func(t *Table, separator string) string {
// 			str := ""
// 			for _, f := range t.GetOrderedCols() {
// 				if len(f.DepFKs) == 0 && len(f.ColName) > 0 {
// 					str = str + f.desc.GetName() + separator
// 				}
// 			}
// 			return strings.TrimSuffix(str, separator)
// 		},
// 		"getFieldName": func(f *field) string {
// 			return f.desc.GetName()
// 		},
// 		"getColumnName": func(f *field) string {
// 			return f.ColName
// 		},
// 	}).Parse(insertTpl)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return tpl
// }
// func (m *Table) Inserter(g Printer) {
// 	err := LoadInsertTemplate().Execute(g, m)
// 	if err != nil {
// 		panic(err)
// 	}
// }
