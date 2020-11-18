package sqlgen

import (
	"strings"
	"text/template"
)

func (m *Table) ConfigStructs(g Printer) {
	err := LoadConfigStructTemplate(g).Execute(g, m)
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

func GetTemplateFuns(p Printer) template.FuncMap {
	TplFuncs["Store"] = func() string {
		return p.StoreName()
	}
	return TplFuncs
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
	"getFieldName":     GetFieldName,
	"getFullFieldName": getFullFieldName,
	"getColumnName": func(f *field) string {
		return f.ColName
	},
	"IsRepeated":          IsRepeated,
	"GetInsertFieldNames": GetInsertFieldNames,
	"GetInsertColNames":   GetInsertColNames,
}

func getFullFieldName(f *field) string {
	table, ok := GetTM().GetTableByTableName(f.Table.Name)
	if ok {
		for _, c := range table.Cols {
			if f.DbfkField == c.ColName {
				return f.desc.GetName() + "." + c.desc.GetName()
			}
		}
	}
	return f.desc.GetName()

}
func GetInsertFieldNames(t *Table, separator string) string {
	str := ""
	for _, f := range t.GetOrderedCols() {
		if f.PK != PK_AUTO && !f.desc.IsRepeated() && len(f.ColName) > 0 {
			if f.desc.IsMessage() {
				tbl, ok := GetTM().GetTableByTableName(f.dbfkTable)
				if !ok {
					continue
				}
				fld, ok := tbl.GetColumnByMessageName(*f.FK.Remote.desc.Name)
				if !ok {
					continue
				}
				str = str + GetFieldName(f) + "." + GetFieldName(fld) + separator
			} else {
				str = str + GetFieldName(f) + separator
			}
		}
	}
	return strings.TrimSuffix(str, separator)
}

func GetInsertColNames(t *Table, separator string) string {
	str := ""
	for _, f := range t.GetOrderedCols() {
		if f.PK != PK_AUTO && !f.desc.IsRepeated() && len(f.ColName) > 0 {
			str = str + f.ColName + separator
		}
	}
	return strings.TrimSuffix(str, separator)
}
func IsRepeated(f *field) bool {
	return f.desc.IsRepeated()
}
func GetFieldName(f *field) string {
	return f.desc.GetName()
}
