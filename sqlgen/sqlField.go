package sqlgen

import (
	"fmt"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
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
	PK        PK
	needQuery bool
	FK        []fieldFK
}

type fieldFK struct {
	message  *generator.Descriptor
	Target   *field
	Source   *field
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
	// if f.desc.IsMessage() {
	// 	// need to check if JSON?
	// 	f.needQuery = true
	// }
	// if f.desc.IsRepeated() {
	// 	// need to check if JSON?
	// 	f.needQuery = true
	// }
	f.setColName()
	f.setPK()
	// f.setFK()
	return f
}
func (f *field) setColName() {
	v, err := proto.GetExtension(f.desc.Options, E_Dbcol)
	if err == nil && v != nil {
		f.ColName = *(v.(*string))
	}
}
func (f *field) setPK() {
	v, err := proto.GetExtension(f.desc.Options, E_Dbpk)
	if err == nil && v != nil {
		f.PK = *(v.(*PK))
	} else {
		f.PK = PK_NONE
	}
}

func (f *field) setFK(tm *TableMessages) {
	// f.FK.relation = FK_NONE
	v, err := proto.GetExtension(f.desc.Options, E_Dbfk)
	if err == nil && v != nil {
		fkArr := strings.Split(*(v.(*string)), ".")
		t, ok := tm.GetTableByTableName(fkArr[0])
		if !ok {
			panic(fmt.Sprintf("Table %s not found", fkArr[0]))
			return
		}
		targetField, ok := t.GetColumnByColumnName(fkArr[1])
		if !ok {
			panic(fmt.Sprintf("Column %s not found", fkArr[1]))
			return
		}
		newRelation := fieldFK{
			message: f.Table.desc,
			Target:  targetField,
			Source:  f,
		}
		if targetField.desc.IsRepeated() {
			newRelation.relation = FK_RELATION_MANY_ONE
		} else {
			newRelation.relation = FK_RELATION_ONE_ONE
		}
		f.FK = append(f.FK, newRelation)
	}
}
