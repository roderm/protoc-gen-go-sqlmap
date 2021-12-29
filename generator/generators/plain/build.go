package plain

import (
	"strings"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen"
	"google.golang.org/protobuf/proto"
)

func (p *SqlGenerator) loadTables() {
	for _, protoFile := range p.plugin.Files {
		for _, msg := range protoFile.Messages {
			ext, ok := proto.GetExtension(msg.Desc.Options(), sqlgen.E_Table).(*sqlgen.Table)
			if !ok || ext == nil {
				continue
			}
			jsonb, ok := proto.GetExtension(msg.Desc.Options(), sqlgen.E_Jsonb).(*bool)
			if !ok || jsonb == nil {
				jsonb = &[]bool{false}[0]
			}
			p.tables.MessageTables[string(msg.Desc.Name())] = &types.Table{
				GoPackageImport: string(protoFile.GoImportPath),
				GoPackageName:   string(protoFile.GoPackageName),
				Engine:          "postgres",
				StoreName:       p.StoreName(protoFile.GeneratedFilenamePrefix),
				MsgName:         string(msg.Desc.Name()),
				Name:            ext.GetName(),
				JSONB:           *jsonb,
				Create:          hasOperation(ext, sqlgen.OPERATION_C),
				Read:            hasOperation(ext, sqlgen.OPERATION_R),
				Update:          hasOperation(ext, sqlgen.OPERATION_U),
				Delete:          hasOperation(ext, sqlgen.OPERATION_D),
				Cols:            make(map[string]*types.Field),
				Imports:         make(map[string]string),
			}
		}
	}
}

func (p *SqlGenerator) loadFields() {
	for _, protoFile := range p.plugin.Files {
		for _, msg := range protoFile.Messages {
			for _, f := range msg.Fields {
				ext, ok := proto.GetExtension(f.Desc.Options(), sqlgen.E_Col).(*sqlgen.Column)
				if !ok || ext == nil {
					continue
				}

				field := &types.Field{
					ColName:    ext.GetName(),
					Table:      p.tables.MessageTables[string(msg.Desc.Name())],
					MsgName:    strings.Title(string(f.Desc.Name())),
					Repeated:   f.Desc.IsList(),
					IsMessage:  f.Desc.Message() != nil,
					Extensions: make(map[string]interface{}),
					PK:         ext.GetPk(),
					Order:      f.Desc.Index(),
					DbfkField:  ext.GetFk(),
					Type:       f.Desc.Kind().String(),
					Oneof: func() string {
						if f.Desc.ContainingOneof() != nil {
							return string(f.Desc.ContainingOneof().Name())
						}
						return ""
					}(),
				}

				p.tables.MessageTables[string(msg.Desc.Name())].Cols[string(f.Desc.Name())] = field
			}
		}
	}
}

func (p *SqlGenerator) loadDependencies() {
	for _, tbl := range p.tables.MessageTables {
		for _, f := range tbl.Cols {
			if f.DbfkField != "" {
				fk := strings.Split(f.DbfkField, ".")
				f.AddForeignKey(p.tables, fk[0], fk[1])
			}

		}
	}
}
