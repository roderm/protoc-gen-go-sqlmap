package types

import (
	"fmt"

	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

type TableRepo []*Table

func (r TableRepo) GetByName(name string) (*Table, bool) {
	for _, tbl := range r {
		if tbl.GetMessageName() == name {
			return tbl, true
		}
	}
	return nil, false
}

type Table struct {
	Def     *sqlmapv1.Table
	File    *protogen.File
	Msg     *protogen.Message
	columns []*Column
}

func NewTableFromDescriptor(f *protogen.File, msg *protogen.Message) (*Table, error) {
	ext := proto.GetExtension(msg.Desc.Options(), sqlmapv1.E_Table).(*sqlmapv1.Table)
	if ext == nil {
		return nil, fmt.Errorf("not defined")
	}
	table := &Table{
		Def:  ext,
		File: f,
		Msg:  msg,
	}
	for _, f := range msg.Fields {
		col, err := NewColumn(table, f)
		if err == nil {
			table.columns = append(table.columns, col)
		}
	}
	return table, nil
}

func (t *Table) GetMessageName() string {
	return string(t.Msg.Desc.Name())
}

func (t *Table) GetTableName() string {
	if t.Def.Name != nil {
		return *t.Def.Name
	}
	return t.GetMessageName()
}

func (t *Table) GetColumns() []*Column {
	return t.columns
}

func (t *Table) GetColumnByFieldName(name string) (*Column, bool) {
	for _, col := range t.GetColumns() {
		if col.GetFieldname() == name {
			return col, true
		}
	}
	return nil, false
}

func (t *Table) GetForeignKeys() []*sqlmapv1.ForeignKeyDefinition {
	fks := []*sqlmapv1.ForeignKeyDefinition{}
	for _, fk := range t.Def.ForeignKeys {
		fks = append(fks, fk)
	}
	for _, c := range t.GetColumns() {
		if c.Def.ForeignKey != nil {
			fks = append(fks, &sqlmapv1.ForeignKeyDefinition{
				Fieldnames: []string{c.Def.GetFieldname()},
				To:         c.Def.ForeignKey,
			})
		}
	}
	return fks
}
