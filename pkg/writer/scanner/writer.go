package scanner

import (
	"html/template"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
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

// Scan reads cols off r, which is satisfied by both *sql.Row and *sql.Rows.
func (n *{{ .Name }}Result) Scan(cols []string, r interface{ Scan(dest ...any) error }) error {
	values := make([]any, len(cols))
	for i, col := range cols {
	switch col {
	{{- range .Columns }}
	case "{{ .FieldName }}":
		{{- if not .IsMessage }}
		values[i] = &n.{{ .MessageName }}
		{{- else }}
		values[i] = &n.fk_{{ .FieldName }}_id
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
	s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "",
		GoImportPath: "fmt",
	})
	tbl := Table{Name: table.GetMessageName()}
	for _, c := range table.GetColumns() {
		tbl.Columns = append(tbl.Columns, &Column{
			IsPrimaryKey: c.Def.GetPk() != schemav1.PK_PK_UNSPECIFIED,
			MessageName:  c.GetName(),
			FieldName:    c.GetFieldname(),
			IsMessage:    c.Field.Desc.Kind() == protoreflect.MessageKind && c.Def.ForeignKey != nil,
			FK:           c.Def.GetForeignKey(),
		})
	}
	return tpl.Execute(s.o, tbl)
}
