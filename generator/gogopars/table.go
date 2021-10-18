package gogopars

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

func NewTable(msg *generator.Descriptor) *types.Table {
	tableName, err := proto.GetExtension(msg.Options, sqlgen.E_Table)
	if err == nil || tableName != nil {
		pt := *(tableName.(*sqlgen.Table))
		tbl := &types.Table{
			// desc:  msg,
			Cols:    make(map[string]*types.Field),
			Name:    pt.GetName(),
			MsgName: msg.GetName(),
			JSONB:   false,
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
		return &types.Table{
			MsgName: msg.GetName(),
			// desc:  msg,
			JSONB: *(JSONB.(*bool)),
		}
	}
	return nil
}

func NewField(table *types.Table, desc *descriptor.FieldDescriptorProto) *types.Field {
	f := &types.Field{
		Table:      table,
		MsgName:    desc.GetName(),
		PK:         sqlgen.PK_NONE,
		Type:       desc.GetType().String(),
		Extensions: make(map[string]interface{}),
		IsMessage:  desc.IsMessage(),
		Order:      int(*desc.Number),
	}
	for _, ext := range []*proto.ExtensionDesc{
		sqlgen.E_Dbcol,
		sqlgen.E_Dbpk,
		sqlgen.E_Dbfk,
	} {
		v, err := proto.GetExtension(desc.Options, ext)
		if err == nil && v != nil {
			f.Extensions[ext.Name] = v
		}
	}
	return f
}

func NewTableMessages(messages []*generator.Descriptor) *types.TableMessages {
	tableMessageStore := &types.TableMessages{
		MessageTables: make(map[string]*types.Table),
	}

	// load tables:
	for _, m := range messages {
		tbl := NewTable(m)
		if tbl != nil {
			tbl.Cols = make(map[string]*types.Field)
			tableMessageStore.MessageTables[m.GetName()] = tbl
			for _, f := range m.Field {
				field := NewField(tbl, f)
				ext, ok := field.Extensions[sqlgen.E_Dbcol.Name]
				if ok && ext != nil {
					field.ColName = *(ext.(*string))
				}
				pk, ok := field.Extensions[sqlgen.E_Dbpk.Name]
				if ok {
					field.PK = *(pk.(*sqlgen.PK))
				}
				tbl.Cols[f.GetName()] = field
			}
		}
	}
	for _, t := range tableMessageStore.MessageTables {
		if t.JSONB {
			continue
		}
		for _, c := range t.Cols {
			v, ok := c.Extensions[sqlgen.E_Dbfk.Name]
			if ok {
				dbfk := *(v.(*string))
				fkArr := strings.Split(dbfk, ".")
				if len(fkArr) != 2 {
					continue
				}
				c.AddForeignKey(tableMessageStore, fkArr[0], fkArr[1])
			}
		}
	}
	return tableMessageStore
}
