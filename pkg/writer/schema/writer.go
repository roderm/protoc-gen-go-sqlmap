package schema

import (
	"fmt"
	"html/template"
	"strings"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"google.golang.org/protobuf/compiler/protogen"
)

// dialects is the set of SQL dialects every generated column carries a
// SchemaType entry for, so one generated file serves any of them at runtime.
var dialects = []string{"mysql", "postgres", "sqlite3"}

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
		Nullable: {{ $proto }}.Bool({{ .Nullable }}),
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
	Name     string // Go identifier suffix for the ColumnDef_ var, e.g. "Id"
	SQLName  string // actual SQL column name, e.g. "simple_id"
	Type     map[string]string
	Pk       string
	Nullable bool
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
	return s.Tables(s.repo.ForFile(protoFile)...)
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
	}

	for _, c := range table.GetColumns() {
		col := &Column{
			Name:     c.GetName(),
			SQLName:  c.GetFieldname(),
			Pk:       pkGoIdent(c.Def.GetPk()),
			Nullable: c.IsNullable(),
			Type:     make(map[string]string, len(dialects)),
		}
		for _, d := range dialects {
			t, err := c.GetSqlType(s.repo, d)
			if err != nil {
				return fmt.Errorf("table %q column %q (%s): %w", table.GetTableName(), c.GetName(), d, err)
			}
			col.Type[d] = t
		}
		tbl.Columns = append(tbl.Columns, col)
	}

	for _, fk := range table.GetForeignKeys() {
		refTable, ok := s.repo.GetByName(fk.To.GetEntity())
		if !ok {
			return fmt.Errorf("table %q: foreign key references unknown entity %q", table.GetTableName(), fk.To.GetEntity())
		}
		external := refTable.File.GoImportPath != table.File.GoImportPath

		out := &ForeignKey{
			ReferencedTable: s.colDefIdent(refTable, fmt.Sprintf("%sTable", refTable.GetMessageName()), external),
			OnDelete:        onDeleteGoIdent(fk.GetOnDelete()),
		}
		for _, name := range fk.GetFieldnames() {
			f, ok := table.GetColumnByFieldName(name)
			if !ok {
				return fmt.Errorf("table %q: foreign key column %q not found", table.GetTableName(), name)
			}
			out.Columns = append(out.Columns, fmt.Sprintf("%sColumnDef_%s", table.GetMessageName(), f.GetName()))
		}
		for _, name := range fk.To.GetFieldnames() {
			f, ok := refTable.GetColumnByFieldName(name)
			if !ok {
				return fmt.Errorf("table %q: referenced column %q not found on %q", table.GetTableName(), name, refTable.GetTableName())
			}
			ident := fmt.Sprintf("%sColumnDef_%s", refTable.GetMessageName(), f.GetName())
			out.ReferencedColumns = append(out.ReferencedColumns, s.colDefIdent(refTable, ident, external))
		}
		tbl.ForeignKeys = append(tbl.ForeignKeys, out)
	}

	return tpl.Execute(s.o, tbl)
}

// colDefIdent qualifies ident with refTable's import path when the referenced
// table lives in another Go package, so cross-file foreign keys compile.
func (s *SchemaWriter) colDefIdent(refTable *types.Table, ident string, external bool) string {
	if !external {
		return ident
	}
	return s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       ident,
		GoImportPath: refTable.File.GoImportPath,
	})
}
