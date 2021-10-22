package types

import (
	"fmt"

	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen/v1"
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
	// desc      *descriptor.FieldDescriptorProto
	Table     *Table
	ColName   string
	Type      string
	MsgName   string
	PK        sqlgen.PK
	needQuery bool
	// dbfk      string
	DbfkField  string
	DbfkTable  string
	FK         FieldFK
	Order      int
	Repeated   bool
	IsMessage  bool
	Extensions map[string]interface{}
	Column     sqlgen.Column
}

type FieldFK struct {
	// message  *generator.Descriptor
	Remote   *Field
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

func (f *Field) AddForeignKey(tm *TableMessages, table string, field string) {
	t, ok := tm.GetTableByTableName(table)
	if !ok {
		panic(fmt.Sprintf("Table '%s' not found", table))
	}
	remoteField, ok := t.GetColumnByColumnName(field)
	if !ok {
		panic(fmt.Sprintf("Column '%s' not found on table '%s'", field, table))
	}
	if f.FK.Remote != nil {
		panic(fmt.Sprintf("Remote is already set %s", field))
	}
	if remoteField.Table == nil {
		panic(fmt.Sprintf("Table is null! %s", remoteField.MsgName))
	}
	if f.Table == nil {
		panic(fmt.Sprintf("Table is null! %s", f.MsgName))
	}
	relation := FK_RELATION_ONE_ONE
	if f.Repeated {
		relation = FK_RELATION_MANY_ONE
	}
	f.FK = FieldFK{
		Remote:   remoteField,
		Relation: int8(relation),
	}
	f.DbfkTable = table
	f.DbfkField = field

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
