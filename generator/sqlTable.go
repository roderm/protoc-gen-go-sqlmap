package generator

import (
	"sort"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

var TableMessageStore *TableMessages

func GetTM() *TableMessages {
	if TableMessageStore == nil {
		panic("Tm not loaded")
	}
	return TableMessageStore
}

type TableMessages struct {
	messageTables map[string]*Table
}

type Table struct {
	Name   string
	desc   *generator.Descriptor
	Cols   map[string]*field
	JSONB  bool
	Create bool
	Read   bool
	Update bool
	Delete bool
}

func (tm *Table) GetColumnByMessageName(message string) (*field, bool) {
	tbl, ok := tm.Cols[message]
	return tbl, ok
}

func (tm *Table) GetColumnByColumnName(message string) (*field, bool) {
	for _, tbl := range tm.Cols {
		if tbl.ColName == message {
			return tbl, true
		}
	}
	return nil, false
}

func (t *Table) GetOrderedCols() []*field {
	r := make(map[int]*field)
	result := []*field{}
	indexes := []int{}
	for _, f := range t.Cols {
		indexes = append(indexes, int(f.desc.GetNumber()))
		r[int(f.desc.GetNumber())] = f
	}
	sort.Ints(indexes)
	for _, i := range indexes {
		result = append(result, r[i])
	}
	return result
}

func NewTableMessages(messages []*generator.Descriptor) *TableMessages {
	TableMessageStore = &TableMessages{
		messageTables: make(map[string]*Table),
	}
	TableMessageStore.loadTables(messages)
	TableMessageStore.loadTableFields()
	TableMessageStore.loadTableFieldFKs()
	return TableMessageStore
}

func (tm *TableMessages) loadTables(messages []*generator.Descriptor) error {
	for _, m := range messages {
		tbl := NewTable(m)
		if tbl != nil {
			tm.messageTables[m.GetName()] = tbl
		}
	}
	return nil
}

func (tm *TableMessages) loadTableFields() {
	for _, t := range tm.messageTables {
		if t.JSONB {
			continue
		}
		t.Cols = make(map[string]*field)
		for _, f := range t.desc.Field {
			t.Cols[f.GetName()] = NewField(t, f)
		}
	}
}

func (tm *TableMessages) loadTableFieldFKs() {
	for _, t := range tm.messageTables {
		if t.JSONB {
			continue
		}
		for _, c := range t.Cols {
			c.setFK(tm)
		}
	}
}

func (tm *TableMessages) GetTableByMessageName(tableName string) (*Table, bool) {
	for _, tbl := range tm.messageTables {
		if tbl.desc.GetName() == tableName {
			return tbl, true
		}
	}
	return nil, false
}

func (tm *TableMessages) GetTableByTableName(tableName string) (*Table, bool) {
	for _, tbl := range tm.messageTables {
		if tbl.Name == tableName {
			return tbl, true
		}
	}
	return nil, false
}

func NewTable(msg *generator.Descriptor) *Table {
	tableName, err := proto.GetExtension(msg.Options, sqlgen.E_Table)
	if err == nil || tableName != nil {
		pt := *(tableName.(*sqlgen.Table))
		tbl := &Table{
			desc:  msg,
			Cols:  make(map[string]*field),
			Name:  pt.GetName(),
			JSONB: false,
		}
		for _, o := range pt.GetCrud() {
			switch o {
			case sqlgen.OPERATION_C:
				tbl.Create = true
			case sqlgen.OPERATION_R:
				tbl.Read = true
			case sqlgen.OPERATION_U:
				tbl.Update = true
			case sqlgen.OPERATION_D:
				tbl.Delete = true
			}
		}
		return tbl
	}
	JSONB, err := proto.GetExtension(msg.Options, sqlgen.E_Jsonb)
	if err == nil || JSONB != nil {
		return &Table{
			desc:  msg,
			JSONB: *(JSONB.(*bool)),
		}
	}
	// tbl.loadFields()
	return nil
}

// func (m *Table) loadFields() {
// 	// TODO: Load
// 	for _, f := range m.desc.Field {
// 		if f.IsRepeated() {
// 			// FK?
// 			mt := proto.MessageType(f.GetTypeName())
// 		}
// 		switch f.GetType() {
// 		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
// 			// Is foreign key or JSON
// 		case descriptor.FieldDescriptorProto_TYPE_BOOL:
// 		}
// 		m.Cols = append(m.Cols, NewField(m, f))

// 	}
// }
func (m Table) GetPKs() []*field {
	fields := []*field{}
	for _, f := range m.Cols {
		if f.PK != sqlgen.PK_NONE {
			fields = append(fields, f)
		}
	}
	return fields
}

func (m *Table) Structs(g Printer) {
	g.P(`type ` + m.Name + `Store struct {
		db  *sql.DB
}`)
	g.P(`type sql` + m.Name + `Array []*` + m.Name)
	g.P(`type sql` + m.Name + ` struct {
		` + m.Name + `
		tmpId string
	}`)
}
