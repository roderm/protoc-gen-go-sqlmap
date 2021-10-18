package writer

import (
	"strings"
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

var TableMessageStore *types.TableMessages

func WriteConfigStructs(g Printer, m *types.Table) {
	err := LoadConfigStructTemplate(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
func GetPK(t *types.Table) *types.Field {
	for _, c := range t.Cols {
		if c.PK != sqlgen.PK_NONE {
			return c
		}
	}
	return nil
}

func SubQueries(t *types.Table) []*types.Field {
	foreignKeys := []*types.Field{}
	for _, f := range t.Cols {
		// && f.desc.IsMessage()
		if f.FK.Remote != nil && f.FK.Remote.Table.Read {
			foreignKeys = append(foreignKeys, f)
		}
	}
	return foreignKeys
}

func GetType(f *types.Field) string {
	switch f.Type {
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
	return f.Type
}

func GetTemplateFuns(p Printer) template.FuncMap {
	TplFuncs["Store"] = func() string {
		return p.StoreName()
	}
	return TplFuncs
}

var TplFuncs = template.FuncMap{
	"SubQueries": SubQueries,
	"GetPKCol": func(t *types.Table) string {
		return GetPK(t).ColName
	},
	"GetPKName": func(t *types.Table) string {
		pk := GetPK(t)
		if pk == nil {
			return "interface{}"
		}
		return pk.MsgName
	},
	"GetPKType": func(t *types.Table) string {
		pk := GetPK(t)
		if pk == nil {
			return "interface{}"
		}
		return GetType(pk)
	},
	"IsReverseFK": func(fk *types.Field) bool {
		return false
	},
	"MessageName": func(t *types.Table) string {
		return t.MsgName
	},
	"TableName": func(t *types.Table) string {
		return t.Name
	},
	"getColumnNames": func(t *types.Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Remote == nil || !f.Repeated) && len(f.ColName) > 0 {
				str = str + f.ColName + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldNames": func(t *types.Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Remote == nil || !f.Repeated) && len(f.ColName) > 0 {
				str = str + f.MsgName + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldName":     GetFieldName,
	"getFullFieldName": getFullFieldName,
	"getColumnName": func(f *types.Field) string {
		return f.ColName
	},
	"IsRepeated":          IsRepeated,
	"GetInsertFieldNames": GetInsertFieldNames,
	"GetInsertColNames":   GetInsertColNames,
	"getInsertFields":     getInsertFields,
}

func getInsertFields(t *types.Table) []*types.Field {
	cols := []*types.Field{}
	inCols := func(new *types.Field) bool {
		for _, c := range cols {
			if new == c {
				return true
			}
		}
		return false
	}
	for _, f := range t.GetOrderedCols() {
		if !inCols(f) && (f.PK == sqlgen.PK_NONE) && f.FK.Remote == nil && len(f.ColName) > 0 {
			cols = append(cols, f)
		}
	}
	return cols
}

func getFullFieldName(f *types.Field) string {
	// table, ok := GetTM().GetTableByTableName(f.Table.Name)
	// if ok {
	// 	for _, c := range table.Cols {
	// 		if f.DbfkField == c.ColName && c.PK != sqlgen.PK_NONE {
	// 			return f.desc.GetName() + "." + c.desc.GetName()
	// 		}
	// 	}
	// }
	return f.MsgName

}
func GetInsertFieldNames(t *types.Table, separator string) string {
	str := ""
	cols := []string{}
	inCols := func(new string) bool {
		for _, c := range cols {
			if new == c {
				return true
			}
		}
		return false
	}
	for _, f := range t.GetOrderedCols() {
		if f.PK == sqlgen.PK_MAN || (f.PK != sqlgen.PK_AUTO && !f.Repeated && len(f.ColName) > 0 && !inCols(f.ColName)) {
			if f.IsMessage {
				tbl, ok := TableMessageStore.GetTableByTableName(f.DbfkTable)
				if !ok {
					continue
				}
				fld, ok := tbl.GetColumnByMessageName(f.FK.Remote.MsgName)
				if !ok {
					continue
				}

				str = str + GetFieldName(f) + "." + GetFieldName(fld) + separator
			} else {
				str = str + GetFieldName(f) + separator
				cols = append(cols, f.ColName)
			}
		}
	}
	return strings.TrimSuffix(str, separator)
}

func GetInsertColNames(t *types.Table, separator string) string {
	str := ""
	cols := []string{}
	inCols := func(new string) bool {
		for _, c := range cols {
			if new == c {
				return true
			}
		}
		return false
	}
	for _, f := range t.GetOrderedCols() {
		if f.PK == sqlgen.PK_MAN || (f.PK != sqlgen.PK_AUTO && !f.Repeated && len(f.ColName) > 0 && !inCols(f.ColName)) {
			cols = append(cols, f.ColName)
			str = str + f.ColName + separator
		}
	}
	return strings.TrimSuffix(str, separator)
}
func IsRepeated(f *types.Field) bool {
	return f.Repeated
}
func GetFieldName(f *types.Field) string {
	return f.MsgName
}
