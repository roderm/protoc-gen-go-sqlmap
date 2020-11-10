package sqlgen

import (
	"strings"
	"text/template"
)

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

func SubQueries(t *Table) []*field {
	foreignKeys := []*field{}
	for _, f := range t.Cols {
		// foreignKeys = append(foreignKeys, f.DepFKs...)
		if f.FK.Remote != nil {
			foreignKeys = append(foreignKeys, f)
		}
	}
	return foreignKeys
}

func GetType(f *field) string {
	switch f.desc.GetType().String() {
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
	return f.desc.GetType().String()
}

var TplFuncs = template.FuncMap{
	"SubQueries": SubQueries,
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
		return GetType(pk)
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
	"IsReverseFK": func(pk *field) bool {
		return false
	},
	// "IsReverseFK": func(remote *Table, data *Table) bool {
	// 	for _, f := range data.Cols {
	// 		if remote.Name == f.dbfkTable {
	// 			if f.desc.IsMessage() {
	// 				mt := strings.Split(f.desc.GetTypeName(), ".")
	// 				return mt[len(mt)-1] == remote.desc.GetName()
	// 			}
	// 		}
	// 	}
	// 	return false
	// },
	"GetRemoteColname": func(remote *Table, data *Table) string {
		for _, f := range data.Cols {
			if remote.Name == f.dbfkTable {
				for _, rf := range remote.Cols {
					if rf.ColName == f.DbfkField {
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
	"GetRemoteListName":  GetRemoteListName,
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
	"getFKMessages": func(t *Table) map[*field]*fieldFK {
		res := make(map[*field]*fieldFK)
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
			if (f.FK.Target == nil || !f.desc.IsRepeated()) && len(f.ColName) > 0 {
				str = str + f.ColName + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldNames": func(t *Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Target == nil || !f.desc.IsRepeated()) && len(f.ColName) > 0 {
				str = str + f.desc.GetName() + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldName": GetFieldName,
	"getFullFieldName": func(f *field) string {
		// if f.desc.IsMessage() {
		table, ok := GetTM().GetTableByTableName(f.Table.Name)
		if !ok {
			return "not_found_" + f.desc.GetName()
		}
		for _, c := range table.Cols {
			if f.DbfkField == c.ColName {
				return f.desc.GetName() + "." + c.desc.GetName()
			}
		}
		return f.desc.GetName() + "_not_found"

	},
	"getColumnName": func(f *field) string {
		return f.ColName
	},
	"getIndexField": getIndexField,
	"getIndexFieldName": func(fk *fieldFK) string {
		return GetFieldName(getIndexField(fk))
	},
	"IsRepeated": IsRepeated,
}

func IsRepeated(f *field) bool {
	return f.desc.IsRepeated()
}
func getIndexField(fk *fieldFK) *field {
	return fk.PKField.FK.Target
	// for _, c := range fk.Target.Table.Cols {
	// 	if c.ColName == fk.PKField.DbfkField {
	// 		return c.FK.PKField
	// 	}
	// }
	// return nil
}
func GetFieldName(f *field) string {
	return f.desc.GetName()
}
func GetRemoteFieldname(remote *Table, data *Table) string {
	for _, f := range data.Cols {
		if remote.Name == f.dbfkTable {
			for _, rf := range remote.Cols {
				if rf.ColName == f.DbfkField {
					return rf.desc.GetName()
				}
			}
		}
	}
	return ""
}
func GetRemoteListName(remote *Table, data *Table) string {
	for _, f := range data.Cols {
		if remote.Name == f.dbfkTable {
			for _, rf := range remote.Cols {
				if rf.ColName == f.DbfkField {
					return rf.desc.GetName()
				}
			}
		}
	}
	return ""
}
