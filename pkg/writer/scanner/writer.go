package scanner

import (
	"html/template"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const templateStr = `
type {{ .Name }}Result struct {
	{{ .Name }}
	{{- range .Columns }}
	{{- if .IsMessage }}
	fk_{{ .FieldName }}_id any
	{{- end }}
	{{- end }}
}

func (n *{{ .Name }}Result) Scan(cols []string, r *sql.Row) error {
	values := make([]any, len(cols))
	for i, col := range cols {
	switch col {
	{{- range .Columns }}
	case "{{ .FieldName }}":
		{{- if not .IsMessage }}
		values[i] = &n.{{ .MessageName }}
		{{- else }}
		// {{ .MessageType }}
		f := new({{ .FK.Entity }})
		values[i] = &f
		{{- end }}
	{{- end }}
	default:
		return fmt.Errorf("unknown column '%s'", col)
	}
	}
	return r.Scan(values...)
}

func (n *{{ .Name }}) GetColValue(col string) any {
	switch col {
	{{- range .Columns }}
	case "{{ .FieldName }}":
		{{- if not .IsMessage }}
		return n.{{ .MessageName }}
		{{- else }}
		return n.{{ .MessageName }}
		{{- end }}
	{{- end }}
	default:
		return nil
	}
}

func Get{{ .Name }}PKColumns() []string {
	return []string{
		{{- range .Columns }}
		{{- if .IsPrimaryKey }}
		"{{ .FieldName }}",
		{{- end }}
		{{- end }}
	}
}

func Get{{ .Name }}Columns() []string {
	return []string{
		{{- range .Columns }}
		"{{ .FieldName }}",
		{{- end }}
	}
}
`

type Column struct {
	IsPrimaryKey bool
	MessageName  string
	FieldName    string
	IsMessage    bool
	FK           *sqlmapv1.ForeignKey
}

type Table struct {
	Name    string
	Columns []*Column
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
	s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "",
		GoImportPath: "database/sql",
	})
	s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "",
		GoImportPath: "fmt",
	})
	tbl := Table{
		Name: table.GetMessageName(),

		Columns: lo.FilterMap(table.GetColumns(), func(c *types.Column, _ int) (*Column, bool) {
			column := &Column{
				IsPrimaryKey: c.Def.Pk != nil && c.Def.Pk != schemav1.PK_PK_UNSPECIFIED.Enum(),
				MessageName:  c.GetName(),
				FieldName:    c.GetFieldname(),
				IsMessage:    c.Field.Desc.Kind() == protoreflect.MessageKind && c.Def.ForeignKey != nil,
				FK:           c.Def.GetForeignKey(),
			}
			return column, true
		}),
	}
	return tpl.Execute(s.o, tbl)
}
