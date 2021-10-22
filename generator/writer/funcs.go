package writer

import (
	"strings"
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen/v1"
)

var TableMessageStore *types.TableMessages

func GetPK(t *types.Table) *types.Field {
	for _, c := range t.Cols {
		if c.PK != sqlgen.PK_PK_UNSPECIFIED {
			return c
		}
	}
	return nil
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
	return TplFuncs
}

var TplFuncs = template.FuncMap{
	"SubQueries": func(t *types.Table) []*types.Field {
		foreignKeys := []*types.Field{}
		for _, f := range t.Cols {
			// && f.desc.IsMessage()
			if f.IsMessage && f.FK.Remote != nil {
				foreignKeys = append(foreignKeys, f)
			}
			// if f.FK.Remote != nil && f.FK.Remote.Table.Read {
			// 	foreignKeys = append(foreignKeys, f)
			// }
		}
		return foreignKeys
	},
	"GetPKCol": func(t *types.Table) string {
		return GetPK(t).ColName
	},
	"GetPKName": func(t *types.Table) string {
		pk := GetPK(t)
		if pk == nil {
			return "interface{}{}"
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
	"getFullFieldName": func(f *types.Field) string {
		table, ok := TableMessageStore.GetTableByTableName(f.Table.Name)
		if ok {
			for _, c := range table.Cols {
				if f.DbfkField == c.ColName && c.PK != sqlgen.PK_PK_UNSPECIFIED {
					return f.MsgName + "." + c.MsgName
				}
			}
		}
		return f.MsgName
	},
	"GetInsertFieldNames": func(t *types.Table, separator string) string {
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
			if f.PK == sqlgen.PK_PK_MAN || (f.PK != sqlgen.PK_PK_AUTO && !f.Repeated && len(f.ColName) > 0 && !inCols(f.ColName)) {
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
	},
	"GetInsertColNames": func(t *types.Table, separator string) string {
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
			if f.PK == sqlgen.PK_PK_MAN || (f.PK != sqlgen.PK_PK_AUTO && !f.Repeated && len(f.ColName) > 0 && !inCols(f.ColName)) {
				cols = append(cols, f.ColName)
				str = str + f.ColName + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getInsertFields": func(t *types.Table) []*types.Field {
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
			if !inCols(f) && (f.PK == sqlgen.PK_PK_UNSPECIFIED) && f.FK.Remote == nil && len(f.ColName) > 0 {
				cols = append(cols, f)
			}
		}
		return cols
	},
}

func GetFieldName(f *types.Field) string {
	return f.MsgName
}
