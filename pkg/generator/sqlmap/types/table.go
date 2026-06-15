package types

import (
	"fmt"

	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

type Table struct {
	def     *sqlmapv1.Table
	file    *protogen.File
	msg     *protogen.Message
	columns []*Column
}

func NewTableFromDescriptor(f *protogen.File, msg *protogen.Message) (*Table, error) {
	ext := proto.GetExtension(msg.Desc.Options(), sqlmapv1.E_Table).(*sqlmapv1.Table)
	if ext == nil {
		return nil, fmt.Errorf("not defined")
	}
	table := &Table{
		def:  ext,
		file: f,
		msg:  msg,
	}
	table.columns = make([]*Column, 0, len(msg.Fields))
	for i, f := range msg.Fields {
		table.columns[i] = NewColumn(table, f)
	}
	return table, nil
}

func (t *Table) GetMessageName() string {
	return string(t.msg.Desc.Name())
}

func (t *Table) GetTableName() string {
	if t.def.Name != nil {
		return *t.def.Name
	}
	return t.GetMessageName()
}

func (t *Table) GetColumns() []*Column {
	return t.columns
}
