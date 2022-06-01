package types

import (
	"sort"
	"strings"

	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

type TableMessages struct {
	MessageTables map[string]*Table
}

type Table struct {
	GoPackageName   string
	GoPackageImport string
	Engine          string
	StoreName       string
	Name            string
	MsgName         string
	Cols            map[string]*Field
	Joins           []*Join
	Config          TableConfig
	Imports         map[string]string
}

type TableConfig struct {
	JSONB  bool
	Create bool
	Read   bool
	Update bool
	Delete bool
}

func (tm *Table) GetColumnByMessageName(message string) (*Field, bool) {
	tbl, ok := tm.Cols[message]
	return tbl, ok
}

func (tm *Table) GetFieldByColumn(message string) (*Field, bool) {
	for _, field := range tm.Cols {
		if field.ColName == message && !field.IsRepeated && field.Oneof == "" && !field.IsMessage {
			return field, true
		}
	}
	for _, field := range tm.Cols {
		if field.ColName == message && !field.IsRepeated && field.Oneof == "" {
			return field, true
		}
	}
	return nil, false
}
func (tm *Table) SimpleFieldByColumn(message string) (*Field, bool) {
	for _, field := range tm.Cols {
		if field.IsRepeated || field.IsMessage {
			continue
		}
		if field.ColName == message {
			return field, true
		}
	}
	return nil, false
}

func (t *Table) GetOrderedCols() []*Field {
	r := make(map[int]*Field)
	result := []*Field{}
	indexes := []int{}
	for _, f := range t.Cols {
		indexes = append(indexes, int(f.Order))
		r[int(f.Order)] = f
	}
	sort.Ints(indexes)
	for _, i := range indexes {
		result = append(result, r[i])
	}
	return result
}

// func NewTableMessages(messages []*generator.Descriptor) *TableMessages {
// 	tableMessageStore = &TableMessages{
// 		MessageTables: make(map[string]*Table),
// 	}
// 	tableMessageStore.loadTables(messages)
// 	// tableMessageStore.loadTableFields()
// 	tableMessageStore.loadTableFieldFKs()
// 	return tableMessageStore
// }

// func (tm *TableMessages) loadTables(messages []*generator.Descriptor) error {
// 	for _, m := range messages {
// 		tbl := NewTable(m)
// 		if tbl != nil {
// 			tm.MessageTables[m.GetName()] = tbl
// 		}
// 	}
// 	return nil
// }

// func (tm *TableMessages) loadTableFields() {
// 	for _, t := range tm.MessageTables {
// 		if t.JSONB {
// 			continue
// 		}
// 		t.Cols = make(map[string]*Field)
// 		for _, f := range t.Cols {
// 			t.Cols[f.GetName()] = NewField(t, f)
// 		}
// 	}
// }

// func (tm *TableMessages) loadTableFieldFKs() {
// 	for _, t := range tm.MessageTables {
// 		if t.JSONB {
// 			continue
// 		}
// 		for _, c := range t.Cols {
// 			c.setFK(tm)
// 		}
// 	}
// }

func (tm *TableMessages) GetTableByMessageName(tableName string) (*Table, bool) {
	for _, tbl := range tm.MessageTables {
		if strings.ToLower(tbl.MsgName) == strings.ToLower(tableName) {
			return tbl, true
		}
	}
	return nil, false
}

func (tm *TableMessages) GetTablesByStoreName(storeName string) []*Table {
	ret := []*Table{}
	for _, tbl := range tm.MessageTables {
		if tbl.StoreName == storeName {
			ret = append(ret, tbl)
		}
	}
	return ret
}
func (tm *TableMessages) GetTableByTableName(tableName string) (*Table, bool) {
	for _, tbl := range tm.MessageTables {
		if tbl.Name == tableName {
			return tbl, true
		}
	}
	return nil, false
}

func (m Table) GetPKs() []*Field {
	Fields := []*Field{}
	for _, f := range m.Cols {
		if f.PK != sqlgen.PK_PK_UNSPECIFIED {
			Fields = append(Fields, f)
		}
	}
	return Fields
}

// func (m *Table) Structs(g Printer) {
// 	g.P(`type ` + m.Name + `Store struct {
// 		db  *sql.DB
// }`)
// 	g.P(`type sql` + m.Name + `Array []*` + m.Name)
// 	g.P(`type sql` + m.Name + ` struct {
// 		` + m.Name + `
// 		tmpId string
// 	}`)
// }
