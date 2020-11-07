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
	dbfk      string
	dbfkField string
	dbfkTable string
	FK        fieldFK
	DepFKs    []*fieldFK
}

type fieldFK struct {
	message  *generator.Descriptor
	Target   *field
	PKField  *field
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
		f.dbfk = *(v.(*string))
		fkArr := strings.Split(f.dbfk, ".")
		if len(fkArr) != 2 {
			return
		}
		f.dbfkTable = fkArr[0]
		f.dbfkField = fkArr[1]
		t, ok := tm.GetTableByTableName(f.dbfkTable)
		if !ok {
			panic(fmt.Sprintf("Table %s not found", f.dbfkTable))
			return
		}
		remoteField, ok := t.GetColumnByColumnName(f.dbfkField)
		if !ok {
			panic(fmt.Sprintf("Column %s not found", f.dbfkField))
			return
		}
		f.FK = fieldFK{
			// message: f.Table.desc,
			PKField: remoteField,
			Target:  f,
		}
		if f.desc.IsRepeated() {
			f.FK.relation = FK_RELATION_MANY_ONE
		} else {
			f.FK.relation = FK_RELATION_ONE_ONE
		}
		remoteField.DepFKs = append(remoteField.DepFKs, &f.FK)
	}
}
