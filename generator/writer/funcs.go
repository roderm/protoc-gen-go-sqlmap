package writer

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen"
)

var TableMessageStore *types.TableMessages

func toPascalCase(in string) string {
	return strcase.ToCamel(in)
}
func GetPK(t *types.Table) *types.Field {
	for _, c := range t.Cols {
		if c.PK != sqlgen.PK_PK_UNSPECIFIED {
			return c
		}
	}
	return nil
}

func getFullFieldName(f *types.Field) string {
	table, ok := TableMessageStore.GetTableByTableName(f.Table.Name)
	if ok {
		for _, c := range table.Cols {
			if f.DbfkField == c.ColName && c.PK != sqlgen.PK_PK_UNSPECIFIED {
				return "Get" + f.MsgName + "()." + strings.Title(c.MsgName)
			}
		}
	}
	return f.MsgName
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

func getInsertFields(t *types.Table) []*types.Field {
	fields := []*types.Field{}
	isPK := func(field *types.Field) bool {
		for _, pk := range t.GetPKs() {
			if pk.ColName == field.ColName {
				return true
			}
		}
		return false
	}
	inCols := func(new *types.Field) bool {
		for _, c := range fields {
			if new.ColName == c.ColName {
				return true
			}
		}
		return false
	}
	for _, f := range t.GetOrderedCols() {
		if f.PK == sqlgen.PK_PK_MAN || (f.PK != sqlgen.PK_PK_AUTO && !f.Repeated && len(f.ColName) > 0 && !inCols(f) && !isPK(f)) {
			fields = append(fields, f)
		}
	}
	return fields
}

var TplFuncs = template.FuncMap{
	"PackagePrefix": func(local *types.Table, remote *types.Table) string {
		if local.GoPackageImport == remote.GoPackageImport {
			return ""
		}
		local.Imports[remote.GoPackageName] = remote.GoPackageImport
		return fmt.Sprintf("%s.", remote.GoPackageName)
	},
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
	"HasPK": func(t *types.Table) bool {
		return GetPK(t) != nil
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
	"GetPKConvert": func(t *types.Table, varName string) string {
		pk := GetPK(t)
		if pk == nil {
			return "interface{}"
		}
		switch GetType(pk) {
		case "int32":
			t.Imports[""] = "encoding/binary"
			return fmt.Sprintf("int32(binary.LittleEndian.Uint32(%s))", varName)
		case "int64":
			t.Imports[""] = "encoding/binary"
			return fmt.Sprintf("int64(binary.LittleEndian.Uint64(%s))", varName)
		}
		return fmt.Sprintf("%s(%s)", GetType(pk), varName)
	},
	"IsReverseFK": func(fk *types.Field) bool {
		return false
	},
	"getColumnNames": func(t *types.Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Remote == nil || !f.Repeated) && len(f.ColName) > 0 && f.Oneof == "" {
				str = str + f.ColName + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFieldNames": func(t *types.Table, separator string) string {
		str := ""
		for _, f := range t.GetOrderedCols() {
			if (f.FK.Remote == nil || !f.Repeated) && len(f.ColName) > 0 && f.Oneof == "" {
				str = str + toPascalCase(f.MsgName) + separator
			}
		}
		return strings.TrimSuffix(str, separator)
	},
	"getFullFieldName": getFullFieldName,
	"GetInsertFieldNames": func(t *types.Table, separator string) string {
		str := ""
		for _, f := range getInsertFields(t) {
			str = str + getFullFieldName(f) + separator
		}
		return strings.TrimSuffix(str, separator)
		// str := ""
		// cols := []string{}
		// inCols := func(new string) bool {
		// 	for _, c := range cols {
		// 		if new == c {
		// 			return true
		// 		}
		// 	}
		// 	return false
		// }
		// for _, f := range t.GetOrderedCols() {
		// 	if f.PK == sqlgen.PK_PK_MAN || (f.PK != sqlgen.PK_PK_AUTO && !f.Repeated && len(f.ColName) > 0 && !inCols(f.ColName)) {
		// 		if f.IsMessage {
		// 			tbl, ok := TableMessageStore.GetTableByTableName(f.DbfkTable)
		// 			if !ok {
		// 				continue
		// 			}
		// 			fld, ok := tbl.GetColumnByMessageName(f.FK.Remote.MsgName)
		// 			if !ok {
		// 				continue
		// 			}

		// 			str = str + GetFieldName(f) + "." + GetFieldName(fld) + separator
		// 		} else {
		// 			str = str + GetFieldName(f) + separator
		// 			cols = append(cols, f.ColName)
		// 		}
		// 	}
		// }
		// return strings.TrimSuffix(str, separator)
	},
	"Title": func(s string) string {
		return strings.Title(s)
	},
	"GetInsertColNames": func(t *types.Table, separator string) string {
		str := ""
		for _, f := range getInsertFields(t) {
			str = str + f.ColName + separator
		}
		return strings.TrimSuffix(str, separator)
	},
	"GetUpdateFields": func(t *types.Table) []*types.Field {
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
