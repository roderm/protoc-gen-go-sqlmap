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
			return "interface{}"
		}
		return pk.desc.GetName()
	},
	"GetPKType": func(t *Table) string {
		pk := GetPK(t)
		if pk == nil {
			return "interface{}"
		}
		return GetType(pk)
	},

	"IsReverseFK": func(pk *field) bool {
		return false
	},

	"MessageName": func(t *Table) string {
		return t.desc.GetName()
	},
	"TableName": func(t *Table) string {
		return t.Name
	},
	"getColumnNames": func(t *Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Remote == nil || !f.desc.IsRepeated()) && len(f.ColName) > 0 {
				str = str + f.ColName + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldNames": func(t *Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Remote == nil || !f.desc.IsRepeated()) && len(f.ColName) > 0 {
				str = str + f.desc.GetName() + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldName": GetFieldName,
	"getFullFieldName": func(f *field) string {
		table, ok := GetTM().GetTableByTableName(f.Table.Name)
		if ok {
			for _, c := range table.Cols {
				if f.DbfkField == c.ColName {
					return f.desc.GetName() + "." + c.desc.GetName()
				}
			}
		}
		return f.desc.GetName()

	},
	"getColumnName": func(f *field) string {
		return f.ColName
	},
	"IsRepeated": IsRepeated,
}

func IsRepeated(f *field) bool {
	return f.desc.IsRepeated()
}
func GetFieldName(f *field) string {
	return f.desc.GetName()
}
