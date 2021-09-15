package generator

import (
	"fmt"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

type fkRelation = int8

const (
	FK_NONE = iota
	FK_RELATION_ONE_MANY
	FK_RELATION_ONE_ONE
	FK_RELATION_MANY_ONE
)

// has [table][col]
var myFields map[string]map[string]*field

func init() {
	myFields = make(map[string]map[string]*field)
}

type field struct {
	desc      *descriptor.FieldDescriptorProto
	Table     *Table
	ColName   string
	PK        sqlgen.PK
	needQuery bool
	dbfk      string
	DbfkField string
	dbfkTable string
	FK        fieldFK
}

type fieldFK struct {
	message  *generator.Descriptor
	Remote   *field
	relation fkRelation
}

func gotField(table, col string) (*field, bool) {
	if table, ok := myFields[table]; ok {
		f, ok := table[col]
		return f, ok
	}
	return nil, false
}
func NewField(table *Table, desc *descriptor.FieldDescriptorProto) *field {
	f := &field{
		Table: table,
		desc:  desc,
	}
	f.setColName()
	f.setPK()
	return f
}
func (f *field) setColName() {
	v, err := proto.GetExtension(f.desc.Options, sqlgen.E_Dbcol)
	if err == nil && v != nil {
		f.ColName = *(v.(*string))
	}
}
func (f *field) setPK() {
	v, err := proto.GetExtension(f.desc.Options, sqlgen.E_Dbpk)
	if err == nil && v != nil {
		f.PK = *(v.(*sqlgen.PK))
	} else {
		f.PK = sqlgen.PK_NONE
	}
}

func (f *field) setFK(tm *TableMessages) {
	v, err := proto.GetExtension(f.desc.Options, sqlgen.E_Dbfk)
	if err == nil && v != nil {
		f.dbfk = *(v.(*string))
		fkArr := strings.Split(f.dbfk, ".")
		if len(fkArr) != 2 {
			return
		}
		f.dbfkTable = fkArr[0]
		f.DbfkField = fkArr[1]
		t, ok := tm.GetTableByTableName(f.dbfkTable)
		if !ok {
			panic(fmt.Sprintf("Table %s not found", f.dbfkTable))
		}
		remoteField, ok := t.GetColumnByColumnName(f.DbfkField)
		if !ok {
			panic(fmt.Sprintf("Column %s not found", f.DbfkField))
		}
		if f.FK.Remote != nil {
			panic(fmt.Sprintf("Remote is already set %s", f.DbfkField))
		}
		f.FK = fieldFK{
			Remote: remoteField,
		}
		if remoteField.Table == nil {
			panic(fmt.Sprintf("Table is null! %s", remoteField.desc.GetName()))
		}
		if f.Table == nil {
			panic(fmt.Sprintf("Table is null! %s", f.desc.GetName()))
		}
		if f.desc.IsRepeated() {
			f.FK.relation = FK_RELATION_MANY_ONE
		} else {
			f.FK.relation = FK_RELATION_ONE_ONE
		}
	}
}
