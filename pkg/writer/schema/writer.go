package schema

import (
	"html/template"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
)

const templateStr = `
var {{ .Name }}Columns []*schema.Column = []*schema.Column{
	{{- range .Columns }}
	&schema.Column{
		Name: "{{ .Name }}",
		Type: "{{ .Type }}",
	},
	{{- end }}
}
var {{ .Name }}Schema *schema.Table = &schema.Table{
	Name: "{{ .TableName }}",
	Columns: {{ .Name }}Columns,
}

`

type Column struct {
	Name string
	Type string
}

type Table struct {
	Name      string
	TableName string
	Columns   []*Column
}

var tpl = template.Must(template.New("schema").Parse(templateStr))

type SchemaWriter struct {
	o       writer.Printer
	dialect string
}

func New(g writer.Printer, dialect string) writer.Writer {
	return &SchemaWriter{o: g, dialect: dialect}
}

func (s *SchemaWriter) Tables(tables ...*types.Table) error {
	s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "schema",
		GoImportPath: "entgo.io/ent/dialect/sql/schema",
	})
	s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "lo",
		GoImportPath: "github.com/samber/lo",
	})
	for _, table := range tables {
		if err := s.table(table); err != nil {
			return err
		}
	}
	return nil
}

func (s *SchemaWriter) table(table *types.Table) error {
	tbl := Table{
		Name:      table.GetTableName(),
		TableName: table.GetTableName(),
		Columns: lo.FilterMap(table.GetColumns(), func(c *types.Column, _ int) (*Column, bool) {
			sqlType, err := c.GetSqlType(s.dialect)
			if err != nil {
				return nil, false
			}
			return &Column{
				Name: c.GetName(),
				Type: sqlType,
			}, true
		}),
	}
	return tpl.Execute(s.o, tbl)
}
