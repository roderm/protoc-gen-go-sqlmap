package types

import (
	"fmt"

	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

type FKRelation = int8

const (
	FK_NONE = iota
	FK_RELATION_ONE_MANY
	FK_RELATION_ONE_ONE
	FK_RELATION_MANY_ONE
)

// has [table][col]
var myFields map[string]map[string]*Field

func init() {
	myFields = make(map[string]map[string]*Field)
}

type Field struct {
	Table   *Table
	ColName string
	Type    string
	MsgName string
	PK      sqlgen.PK
	Oneof   string
	// dbfk      string
	DbfkField string
	DbfkTable string
	FK        FieldFK

	// FkOut []Field
	// FkIn  []Field

	Order       int
	IsRepeated  bool
	IsMessage   bool
	TypeMessage string
	Extensions  map[string]interface{}
	Column      sqlgen.Column
}
type MessageField struct {
	Field
	IsRepeated bool
	Target     Field
}

type FieldFK struct {
	// message  *generator.Descriptor
	ChildOf  *Field
	ParentOf []*Field
	Relation FKRelation
}

func gotField(table, col string) (*Field, bool) {
	if table, ok := myFields[table]; ok {
		f, ok := table[col]
		return f, ok
	}
	return nil, false
}

// func NewField(table *Table, desc *descriptorpb.FieldDescriptorProto) *Field {
// 	f := &Field{
// 		Table:   table,
// 		MsgName: desc.GetName(),
// 	}
// 	f.setColName(desc)
// 	f.setPK(desc)
// 	return f
// }
// func (f *Field) setColName(desc *descriptorpb.FieldDescriptorProto) {
// 	v := proto.GetExtension(desc.Options, sqlgen.E_Dbcol)
// 	if v != nil {
// 		f.ColName = *(v.(*string))
// 	}
// }
// func (f *Field) setPK(desc *descriptorpb.FieldDescriptorProto) {
// 	v, ok := proto.GetExtension(desc.Options, sqlgen.E_Dbpk).(*sqlgen.PK)
// 	if ok && v != nil {
// 		f.PK = *v
// 	} else {
// 		f.PK = sqlgen.PK_PK_UNSPECIFIED
// 	}
// }

func (f *Field) AddForeignKey(tm *TableMessages) {
	if !f.IsMessage {
		return
	}
	t, ok := tm.GetTableByTableName(f.DbfkTable)
	if !ok {
		panic(fmt.Sprintf("Table '%s' not found", f.DbfkTable))
	}
	remoteField, ok := t.GetFieldByColumn(f.DbfkField)
	if !ok {
		panic(fmt.Sprintf("Column '%s' not found on table '%s'", f.DbfkField, f.DbfkTable))
	}
	if f.FK.ChildOf != nil {
		panic(fmt.Sprintf("Remote is already set %s", f.DbfkField))
	}
	if remoteField.Table == nil {
		panic(fmt.Sprintf("Table is null! %s", remoteField.MsgName))
	}
	if f.Table == nil {
		panic(fmt.Sprintf("Table is null! %s", f.MsgName))
	}
	relation := FK_RELATION_ONE_ONE
	if f.IsRepeated {
		relation = FK_RELATION_MANY_ONE
	}
	f.FK = FieldFK{
		ChildOf:  remoteField,
		Relation: int8(relation),
	}

	f.Table.Joins = append(f.Table.Joins, &Join{
		IsRepeated:        f.IsRepeated,
		Target:            f,
		Source:            remoteField,
		TargetMessageName: f.Table.MsgName,
		TargetFieldName:   f.MsgName,
		TargetIsOneOf:     f.Oneof != "",
		TargetOneOfField:  f.Oneof,
		TargetSourceKeyField: func() string {
			ChildIdField, ok := f.Table.GetFieldByColumn(f.ColName)
			if !ok {
				panic(fmt.Errorf("Table %s has no column %s (in ChildIdField for %s)", f.Table.Name, f.ColName, remoteField.MsgName))
			}
			if ChildIdField.IsMessage {
				tbl, ok := tm.GetTableByTableName(ChildIdField.DbfkTable)
				if !ok {
					panic(fmt.Errorf("message %s not loaded", ChildIdField.DbfkTable))
				}
				idField, ok := tbl.GetFieldByColumn(ChildIdField.DbfkField)
				if !ok {
					panic(fmt.Errorf("Table %s has no column %s (in ChildIdField.IsMessage %s)", tbl.Name, ChildIdField.DbfkField, ChildIdField.MsgName))
				}
				if idField.Oneof != "" {
					return fmt.Sprintf("Get%s().%s /* catched oneof %v */", ChildIdField.MsgName, "ID", idField)
				}
				return fmt.Sprintf("Get%s().%s", ChildIdField.MsgName, idField.MsgName)
			} else {
				return ChildIdField.MsgName
			}
		}(),
		SourceTargetKeyField: func() string {
			ParentIdField, ok := remoteField.Table.GetFieldByColumn(f.DbfkField)
			if !ok {
				panic(fmt.Errorf("Table %s has no column %s (in ParentIdField)", remoteField.Table.Name, f.DbfkField))
			}
			if ParentIdField.IsMessage {
				tbl, ok := tm.GetTableByTableName(ParentIdField.DbfkTable)
				if !ok {
					panic(fmt.Errorf("message %s not loaded", ParentIdField.DbfkTable))
				}
				idField, ok := tbl.GetFieldByColumn(ParentIdField.DbfkField)
				if !ok {
					panic(fmt.Errorf("Table %s has no column %s (in ParentIdField.IsMessage)", tbl.Name, ParentIdField.DbfkField))
				}
				if idField.Oneof != "" {
					return fmt.Sprintf("Get%s().%s /* catched oneof */", ParentIdField.MsgName, "ID")
				}
				return fmt.Sprintf("Get%s().%s", ParentIdField.MsgName, idField.MsgName)
			} else {
				return ParentIdField.MsgName
			}
		}(),
		SourcePackagePrefix: func() string {
			if f.Table.GoPackageImport == remoteField.Table.GoPackageImport {
				return ""
			}
			f.Table.Imports[remoteField.Table.GoPackageName] = remoteField.Table.GoPackageImport
			return fmt.Sprintf("%s.", remoteField.Table.GoPackageName)
		}(),
		SourceMessageName: remoteField.Table.MsgName,
		SourceFieldName:   remoteField.MsgName,
		SourceColumnName:  remoteField.ColName,
	})
}

// func (f *Field) setFK(tm *TableMessages) {
// 	v, err := proto.GetExtension(f.desc.Options, sqlgen.E_Dbfk)
// 	if err == nil && v != nil {
// 		dbfk := *(v.(*string))
// 		fkArr := strings.Split(dbfk, ".")
// 		if len(fkArr) != 2 {
// 			return
// 		}
// 		f.dbfkTable = fkArr[0]
// 		f.DbfkField = fkArr[1]
// 		t, ok := tm.GetTableByTableName(f.dbfkTable)
// 		if !ok {
// 			panic(fmt.Sprintf("Table %s not found", f.dbfkTable))
// 		}
// 		remoteField, ok := t.GetColumnByColumnName(f.DbfkField)
// 		if !ok {
// 			panic(fmt.Sprintf("Column %s not found", f.DbfkField))
// 		}
// 		if f.FK.Remote != nil {
// 			panic(fmt.Sprintf("Remote is already set %s", f.DbfkField))
// 		}
// 		f.FK = FieldFK{
// 			Remote: remoteField,
// 		}
// 		if remoteField.Table == nil {
// 			panic(fmt.Sprintf("Table is null! %s", remoteField.desc.GetName()))
// 		}
// 		if f.Table == nil {
// 			panic(fmt.Sprintf("Table is null! %s", f.desc.GetName()))
// 		}
// 		if f.desc.IsRepeated() {
// 			f.FK.relation = FK_RELATION_MANY_ONE
// 		} else {
// 			f.FK.relation = FK_RELATION_ONE_ONE
// 		}
// 	}
// }
