package schema

import (
	"fmt"
	"html/template"
	"strings"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// schemaV1ImportPath is the generated Go package holding SchemaTable/SchemaColumn/
// SchemaForeignKey and the PK/OnDelete enums that this writer emits into.
const schemaV1ImportPath = protogen.GoImportPath("github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1")

// protoImportPath is google.golang.org/protobuf/proto, used for proto.String()
// to populate the schema.v1 messages' proto2 *string fields.
const protoImportPath = protogen.GoImportPath("google.golang.org/protobuf/proto")

const templateStr = `
{{- $pkg := .SchemaPkg }}
{{- $proto := .ProtoPkg }}
var (
{{- $msg := .Name }}
{{- range .Columns }}
	{{ $msg }}ColumnDef_{{ .Name }} = &{{ $pkg }}.SchemaColumn{
		Name: {{ $proto }}.String("{{ .SQLName }}"),
		Type: map[string]string{
			{{ range $k, $v := .Type }}"{{ $k }}": "{{ $v }}",
			{{ end }}
		},
		Pk: {{ $pkg }}.{{ .Pk }}.Enum(),
	}
{{- end }}
)
var {{ $msg }}Table = &{{ $pkg }}.SchemaTable{
	Name: {{ $proto }}.String("{{ .TableName }}"),
	Columns: []*{{ $pkg }}.SchemaColumn{
		{{- range .Columns }}
		{{ $msg }}ColumnDef_{{ .Name }},
		{{- end }}
	},
}

func init() {
{{- $tableName := printf "%sTable" .Name }}
{{- range $fk := .ForeignKeys }}
	{{ $tableName }}.ForeignKeys = append({{ $tableName }}.ForeignKeys, &{{ $pkg }}.SchemaForeignKey{
		Columns: []*{{ $pkg }}.SchemaColumn{
			{{- range  $v := .Columns }}
			{{ $v }},
			{{- end }}
		},
		RefTable:   {{ $fk.ReferencedTable }},
		RefColumns: []*{{ $pkg }}.SchemaColumn{
			{{- range $v := $fk.ReferencedColumns }}
			{{ $v }},
			{{- end }}
		},
		OnDelete: {{ $pkg }}.{{ $fk.OnDelete }}.Enum(),
	})
	{{- end }}
}

`

type Column struct {
	Name    string // Go identifier suffix for the ColumnDef_ var, e.g. "Id"
	SQLName string // actual SQL column name, e.g. "simple_id"
	Type    map[string]string
	Pk      string
}

type Table struct {
	Name        string
	TableName   string
	SchemaPkg   string
	ProtoPkg    string
	Columns     []*Column
	ForeignKeys []*ForeignKey
}

type ForeignKey struct {
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnDelete          string
}

// pkGoIdent returns the generated Go constant identifier (not PK's proto
// enum-value name, which protoc-gen-go doesn't prefix the same way) for pk.
func pkGoIdent(pk schemav1.PK) string {
	switch pk {
	case schemav1.PK_PK_AUTO:
		return "PK_PK_AUTO"
	case schemav1.PK_PK_MAN:
		return "PK_PK_MAN"
	default:
		return "PK_PK_UNSPECIFIED"
	}
}

// onDeleteGoIdent returns the generated Go constant identifier for od.
func onDeleteGoIdent(od schemav1.OnDelete) string {
	switch od {
	case schemav1.OnDelete_ON_DELETE_CASCADE:
		return "OnDelete_ON_DELETE_CASCADE"
	case schemav1.OnDelete_ON_DELETE_SET_NULL:
		return "OnDelete_ON_DELETE_SET_NULL"
	default:
		return "OnDelete_ON_DELETE_UNSPECIFIED"
	}
}

var tpl = template.Must(template.New("schema").Parse(templateStr))

type SchemaWriter struct {
	o    writer.Printer
	repo types.TableRepo
}

func New(g writer.Printer, repo types.TableRepo) writer.Writer {
	return &SchemaWriter{o: g, repo: repo}
}

func (s *SchemaWriter) Write(protoFile *protogen.File) error {
	entities := lo.FilterMap(protoFile.Messages, func(msg *protogen.Message, _ int) (protoreflect.FullName, bool) {
		return msg.Desc.FullName(), true
	})
	tables := lo.Filter(s.repo, func(table *types.Table, _ int) bool {
		return lo.Contains(entities, table.Msg.Desc.FullName())
	})
	if err := s.Tables(tables...); err != nil {
		return err
	}
	return nil
}

func (s *SchemaWriter) Tables(tables ...*types.Table) error {
	for _, table := range tables {
		if err := s.table(table); err != nil {
			return err
		}
	}
	return nil
}

func (s *SchemaWriter) table(table *types.Table) error {
	schemaPkg := strings.TrimSuffix(s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "SchemaTable",
		GoImportPath: schemaV1ImportPath,
	}), ".SchemaTable")
	protoPkg := strings.TrimSuffix(s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "String",
		GoImportPath: protoImportPath,
	}), ".String")

	tbl := Table{
		Name:      table.GetMessageName(),
		TableName: table.GetTableName(),
		SchemaPkg: schemaPkg,
		ProtoPkg:  protoPkg,
		ForeignKeys: lo.Map(table.GetForeignKeys(), func(fk *sqlmapv1.ForeignKeyDefinition, _ int) *ForeignKey {
			refTable, ok := s.repo.GetByName(fk.To.GetEntity())
			if !ok {
				return nil
			}
			tableName := fmt.Sprintf("%sTable", refTable.GetMessageName())
			if refTable.File.GoImportPath != table.File.GoImportPath {
				tableName = s.o.QualifiedGoIdent(protogen.GoIdent{
					GoName:       tableName,
					GoImportPath: refTable.File.GoImportPath,
				})
			}
			return &ForeignKey{
				Columns: lo.Map(fk.GetFieldnames(), func(c string, _ int) string {
					f, ok := table.GetColumnByFieldName(c)
					if !ok {
						return ""
					}
					return fmt.Sprintf("%sColumnDef_%s", table.GetMessageName(), f.GetName())
				}),
				ReferencedTable: tableName,
				ReferencedColumns: lo.Map(fk.To.GetFieldnames(), func(c string, _ int) string {
					f, ok := refTable.GetColumnByFieldName(c)
					if !ok {
						return ""
					}
					name := fmt.Sprintf("%sColumnDef_%s", refTable.GetMessageName(), f.GetName())
					if refTable.File.GoImportPath != table.File.GoImportPath {
						return s.o.QualifiedGoIdent(protogen.GoIdent{
							GoName:       name,
							GoImportPath: refTable.File.GoImportPath,
						})
					}
					return name
				}),
				OnDelete: onDeleteGoIdent(fk.GetOnDelete()),
			}
		}),
		Columns: lo.FilterMap(table.GetColumns(), func(c *types.Column, _ int) (*Column, bool) {
			return &Column{
				Name:    c.GetName(),
				SQLName: c.GetFieldname(),
				Pk:      pkGoIdent(c.Def.GetPk()),
				Type: lo.SliceToMap([]string{"mysql", "postgres", "sqlite3"}, func(d string) (string, string) {
					t, err := c.GetSqlType(s.repo, d)
					if err != nil {
						return d, "unknown[err: " + err.Error() + "]"
					}
					return d, t
				}),
			}, true
		}),
	}
	return tpl.Execute(s.o, tbl)
}
