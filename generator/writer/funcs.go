package writer

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
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
	if f.IsMessage {
		table, ok := TableMessageStore.GetTableByTableName(f.Table.Name)
		if ok {
			for _, c := range table.Cols {
				// && c.PK != sqlgen.PK_PK_UNSPECIFIED
				if f.DbfkField == c.ColName && !c.IsMessage {
					return fmt.Sprintf("Get%s().Get%s()", f.MsgName, toPascalCase(c.MsgName))
				}
			}
		}
	}
	return fmt.Sprintf("Get%s()", f.MsgName)
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

func GetTemplateFuns() template.FuncMap {
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
		if f.PK == sqlgen.PK_PK_MAN || (f.PK != sqlgen.PK_PK_AUTO && !f.IsRepeated && len(f.ColName) > 0 && !inCols(f) && !isPK(f)) {
			fields = append(fields, f)
		}
	}
	return fields
}

func GetPrimaries(t *types.Table) []*types.Field {
	cols := []*types.Field{}
	for _, c := range t.GetOrderedCols() {
		if c.PK != sqlgen.PK_PK_UNSPECIFIED {
			cols = append(cols, c)
		}
	}
	return cols
}

var TplFuncs = template.FuncMap{
	"GetPrimaryBase": func(t *types.Table) int {
		return len(GetPrimaries(t))
	},
	"GetPrimaries": GetPrimaries,
	"GetPrimaryCols": func(t *types.Table) string {
		names := []string{}
		for i, n := range GetPrimaries(t) {
			names = append(names, fmt.Sprintf("\"%s\" = $%d", n.ColName, i+1))
		}
		return strings.Join(names, " AND ")
	},
	"GetPrimaryValues": func(t *types.Table, obj string) string {
		names := []string{}
		for _, n := range GetPrimaries(t) {
			names = append(names, fmt.Sprintf("%s.Get%s()", obj, n.MsgName))
		}
		return strings.Join(names, ", ")
	},
	"PackagePrefix": func(local *types.Table, remote *types.Table) string {
		if local.GoPackageImport == remote.GoPackageImport {
			return ""
		}
		local.Imports[remote.GoPackageName] = remote.GoPackageImport
		return fmt.Sprintf("%s.", remote.GoPackageName)
	},
	"GetMessageFields": func(t *types.Table) []*types.Field {
		fields := []*types.Field{}
		for _, f := range t.Cols {
			if f.IsMessage {
				fields = append(fields, f)
			}
		}
		return fields
	},
	"SubQueries": func(t *types.Table) []*types.Field {
		foreignKeys := []*types.Field{}
		for _, f := range t.Cols {
			if f.IsMessage && f.FK.ChildOf != nil {
				foreignKeys = append(foreignKeys, f)
			}
		}
		sort.Slice(foreignKeys, func(i, j int) bool {
			return foreignKeys[i].MsgName < foreignKeys[j].MsgName
		})
		return foreignKeys
	},
	"HasPK": func(t *types.Table) bool {
		return GetPK(t) != nil
	},
	"GetPKCol": func(t *types.Table) string {
		return GetPK(t).ColName
	},
	"RepeatedFKFieldGetter": func(t *types.Table, remote *types.Field) string {
		if remote.FK.ChildOf != nil {
			for _, c := range remote.FK.ChildOf.Table.Cols {
				if c.DbfkField == remote.ColName {
					for _, f := range c.FK.ChildOf.Table.Cols {
						if !f.IsRepeated && !f.IsMessage && f.ColName == c.DbfkField {
							return fmt.Sprintf("Get%s()", f.MsgName)
						}
					}
				}
			}
		}
		for _, c := range t.Cols {
			if c.ColName == remote.DbfkField {
				return fmt.Sprintf("Get%s()", c.MsgName)
			}
		}
		return fmt.Sprintf("GetUnknown%s()", remote.MsgName)
	},
	"GetParentFieldColName": func(t *types.Table, r *types.Field) string {
		if r.FK.ChildOf != nil {
			for _, c := range r.FK.ChildOf.Table.Cols {
				if c.DbfkField == r.ColName {
					for _, f := range c.FK.ChildOf.Table.Cols {
						if f.ColName == c.DbfkField {
							return f.ColName
						}
					}
				}
			}
		}
		return ""
	},
	"GetParentFieldForChild": func(t *types.Table, r *types.Field) string {
		if r.FK.ChildOf != nil {
			for _, c := range r.FK.ChildOf.Table.Cols {
				if c.DbfkField == r.ColName {
					for _, f := range c.FK.ChildOf.Table.Cols {
						if !f.IsRepeated && !f.IsMessage && f.ColName == c.DbfkField {
							return fmt.Sprintf("Get%s().Get%s()", c.MsgName, f.MsgName)
						}
					}
				}
			}
		}
		for _, c := range t.Cols {
			if c.ColName == r.DbfkField {
				return fmt.Sprintf("Get%s()", c.MsgName)
			}
		}
		return fmt.Sprintf("GetTable().GetField()")
	},
	"GetParentFieldForParent": func(t *types.Table, r *types.Field) string {
		if r.FK.ChildOf != nil {
			for _, c := range r.FK.ChildOf.Table.Cols {
				if c.DbfkField == r.ColName {
					for _, f := range c.FK.ChildOf.Table.Cols {
						if !f.IsRepeated && !f.IsMessage && f.ColName == c.DbfkField {
							return fmt.Sprintf("Get%s()", f.MsgName)
						}
					}
				}
			}
		}
		for _, c := range t.Cols {
			if c.ColName == r.DbfkField {
				return fmt.Sprintf("Get%s()", c.MsgName)
			}
		}
		return fmt.Sprintf("GetTable().GetField()")
	},
	"MessageFKFieldGetter": func(t *types.Table, remote *types.Field) string {
		if remote.IsMessage {
			fieldname := ""
			for _, f := range remote.FK.ChildOf.Table.Cols {
				if f.ColName == remote.DbfkField {
					fieldname = f.MsgName
				}
			}
			return fmt.Sprintf("Get%s().Get%s()", remote.MsgName, fieldname)
		} else {
			return fmt.Sprintf("Get%s()", remote.MsgName)
		}
	},
	"MessageFKField": func(t *types.Table, remote *types.Field) string {
		return fmt.Sprintf("Get%s()", remote.FK.ChildOf.MsgName)
	},
	"MessageFKItemField": func(t *types.Table, remote *types.Field) string {
		for _, c := range t.Cols {
			if c.ColName == remote.DbfkField {
				return fmt.Sprintf("Get%s()", c.MsgName)
			}
		}
		if remote.FK.ChildOf != nil {
			for _, c := range remote.FK.ChildOf.Table.Cols {
				if c.DbfkField == remote.ColName {
					for _, f := range c.FK.ChildOf.Table.Cols {
						if !f.IsRepeated && !f.IsMessage && f.ColName == c.DbfkField {
							return fmt.Sprintf("Get%s()", f.MsgName)
						}
					}
				}
			}
		}
		return fmt.Sprintf("GetUnkown()/* %s => %s */", remote.MsgName, remote.DbfkField)
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
	"getColumnNames": func(t *types.Table, separator string) string {
		cols := []string{}
		for _, f := range t.GetOrderedCols() {
			if (!f.IsRepeated) && len(f.ColName) > 0 && f.Oneof == "" {
				if f.IsMessage && f.FK.ChildOf != nil {
					cols = append(cols, "JSON_BUILD_OBJECT('"+f.FK.ChildOf.ColName+"', \""+f.ColName+"\") AS "+f.ColName)
				} else {
					cols = append(cols, fmt.Sprintf("\"%s\"", f.ColName))
				}
			}
		}
		return strings.Join(cols, separator)
	},
	"getMessageFields": func(t *types.Table) []*types.Field {
		fields := []*types.Field{}
		for _, f := range t.Cols {
			if f.IsMessage {
				fields = append(fields, f)
			}
		}
		return fields
	},
	"getFieldNames": func(t *types.Table, separator string) string {
		cols := []string{}
		for _, f := range t.GetOrderedCols() {
			if f.FK.ChildOf != nil && f.FK.ChildOf.Table.Config.JSONB {
				cols = append(cols, toPascalCase(f.MsgName))
				continue
			}
			if !f.IsMessage && !f.IsRepeated && len(f.ColName) > 0 && f.Oneof == "" {
				cols = append(cols, toPascalCase(f.MsgName))
			}
		}
		for _, f := range t.Joins {
			if f.IsRepeated || f.Target.Oneof != "" {
				continue
			}
			cols = append(cols, fmt.Sprintf("Get%s().%s", f.Target.MsgName, f.SourceTargetKeyField))
		}
		return strings.Join(cols, separator)
	},
	"getFullFieldName": getFullFieldName,
	"GetInsertFieldNames": func(t *types.Table, obj string, separator string) string {
		str := []string{}
		for _, f := range getInsertFields(t) {
			if f.IsMessage {
				str = append(str, fmt.Sprintf("%s.%s", obj, getFullFieldName(f)))
			} else {
				str = append(str, fmt.Sprintf("%s.Get%s()", obj, f.MsgName))
			}
		}
		return strings.Join(str, separator)
	},
	"Title": func(s string) string {
		return toPascalCase(s)
	},
	"GetInsertFields": func(t *types.Table) []*types.Field {
		return getInsertFields(t)
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
			if !inCols(f) && (f.PK == sqlgen.PK_PK_UNSPECIFIED) && f.FK.ChildOf == nil && len(f.ColName) > 0 {
				cols = append(cols, f)
			}
		}
		return cols
	},
}

func GetFieldName(f *types.Field) string {
	return f.MsgName
}
