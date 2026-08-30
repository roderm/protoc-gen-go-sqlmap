package scanner

// text/template, not html/template: HTML escaping mangles quotes in Go source.
import (
	"strings"
	"text/template"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const templateStr = `
{{- $sql := .SQLPkg }}
{{- $ts := .TimestamppbPkg }}
type {{ .Name }}Result struct {
	{{ .Name }}
	{{- range .Columns }}
	{{- if .IsMessage }}
	fk_{{ .FieldName }}_id any
	{{- else if .IsTimestamp }}
	ts_{{ .FieldName }} {{ $sql }}.NullTime
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
		{{- if .IsMessage }}
		values[i] = &n.fk_{{ .FieldName }}_id
		{{- else if .IsTimestamp }}
		values[i] = &n.ts_{{ .FieldName }}
		{{- else }}
		values[i] = &n.{{ .MessageName }}
		{{- end }}
	{{- end }}
	default:
		return fmt.Errorf("unknown column '%s'", col)
	}
	}
	if err := r.Scan(values...); err != nil {
		return err
	}
	{{- range .Columns }}
	{{- if .IsTimestamp }}
	// Left nil when the column is NULL, so an absent timestamp stays absent
	// rather than becoming the zero instant.
	if n.ts_{{ .FieldName }}.Valid {
		n.{{ .MessageName }} = {{ $ts }}.New(n.ts_{{ .FieldName }}.Time)
	}
	{{- end }}
	{{- end }}
	return nil
}

func (n *{{ .Name }}) GetColValue(col string) any {
	switch col {
	{{- range .Columns }}
	case "{{ .FieldName }}":
		{{- if .IsTimestamp }}
		// Handed back as a time.Time, which is what a driver can bind.
		if n.{{ .MessageName }} == nil {
			return nil
		}
		return n.{{ .MessageName }}.AsTime()
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
	IsTimestamp  bool
	FK           *sqlmapv1.ForeignKey
}

type Table struct {
	Name string
	// Empty unless the table has a timestamp column: qualifying an identifier
	// is what adds the import, and an unused one would not compile.
	SQLPkg         string
	TimestamppbPkg string
	Columns        []*Column
}

// timestamppbImportPath provides timestamppb.New for the scanned time.Time.
const timestamppbImportPath = protogen.GoImportPath("google.golang.org/protobuf/types/known/timestamppb")

var tpl = template.Must(template.New("schema").Parse(templateStr))

// pkgAlias resolves the local alias protogen assigned to an import.
func (s *SchemaWriter) pkgAlias(goName string, path protogen.GoImportPath, suffix string) string {
	return strings.TrimSuffix(s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       goName,
		GoImportPath: path,
	}), suffix)
}

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
		if c.IsTimestamp() {
			// Lazy: requesting these is what adds the imports.
			tbl.SQLPkg = s.pkgAlias("NullTime", "database/sql", ".NullTime")
			tbl.TimestamppbPkg = s.pkgAlias("New", timestamppbImportPath, ".New")
		}
		tbl.Columns = append(tbl.Columns, &Column{
			IsPrimaryKey: c.Def.GetPk() != schemav1.PK_PK_UNSPECIFIED,
			IsTimestamp:  c.IsTimestamp(),
			MessageName:  c.GetName(),
			FieldName:    c.GetFieldname(),
			IsMessage:    c.Field.Desc.Kind() == protoreflect.MessageKind && c.Def.ForeignKey != nil,
			FK:           c.Def.GetForeignKey(),
		})
	}
	return tpl.Execute(s.o, tbl)
}
