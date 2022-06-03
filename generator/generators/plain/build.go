package plain

import (
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
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
				Engine:          p.SQLDialect(),
				StoreName:       p.StoreName(protoFile.GeneratedFilenamePrefix),
				MsgName:         strcase.ToCamel(string(msg.Desc.Name())),
				Name:            ext.GetName(),
				Joins:           []*types.Join{},
				Config: types.TableConfig{
					JSONB:  *jsonb,
					Create: hasOperation(ext, sqlgen.OPERATION_C),
					Read:   hasOperation(ext, sqlgen.OPERATION_R),
					Update: hasOperation(ext, sqlgen.OPERATION_U),
					Delete: hasOperation(ext, sqlgen.OPERATION_D),
				},
				Cols:    make(map[string]*types.Field),
				Imports: make(map[string]string),
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
					ColName: ext.GetName(),
					Table:   p.tables.MessageTables[string(msg.Desc.Name())],
					FK: types.FieldFK{
						ParentOf: []*types.Field{},
					},
					MsgName:    strcase.ToCamel(string(f.Desc.Name())),
					IsRepeated: f.Desc.IsList(),
					IsMessage:  f.Desc.Message() != nil,
					Extensions: make(map[string]interface{}),
					PK:         ext.GetPk(),
					Order:      f.Desc.Index(),
					Type:       f.Desc.Kind().String(),
					TypeMessage: func() string {
						if f.Desc.Message() != nil {
							return strcase.ToCamel(string(f.Desc.Message().Name()))
						}
						return ""
					}(),
					Oneof: func() string {
						if f.Desc.ContainingOneof() != nil {
							return strcase.ToCamel(string(f.Desc.ContainingOneof().Name()))
						}
						return ""
					}(),
				}
				fkPath := strings.Split(ext.GetFk(), ".")
				if len(fkPath) == 2 {
					field.DbfkTable = fkPath[0]
					field.DbfkField = fkPath[1]
				} else {
					field.DbfkField = fkPath[0]
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
				f.AddForeignKey(p.tables)
			}

		}
	}
}
