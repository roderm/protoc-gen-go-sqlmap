// Package query emits FieldMask-aware row loading. Related rows are fetched
// with one batched `IN (...)` per relation and stitched in Go rather than
// joined, which would multiply each parent row out once per child.
package query

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/snaerverk/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/snaerverk/protoc-gen-go-sqlmap/pkg/writer"
	"google.golang.org/protobuf/compiler/protogen"
)

const (
	queryImportPath     = protogen.GoImportPath("github.com/snaerverk/protoc-gen-go-sqlmap/pkg/query")
	fieldmaskImportPath = protogen.GoImportPath("google.golang.org/protobuf/types/known/fieldmaskpb")
	contextImportPath   = protogen.GoImportPath("context")
)

const templateStr = `
{{- $q := .QueryPkg }}
{{- $msg := .Name }}
// {{ $msg }}QueryColumns returns the columns to select for mask. Primary keys
// are always included, since relations are stitched on them.
func {{ $msg }}QueryColumns(m *{{ $q }}.Mask, extra ...string) []string {
	cols := make([]string, 0, {{ len .Columns }}+len(extra))
	{{- range .Columns }}
	{{- if .IsPK }}
	cols = append(cols, "{{ .SQLName }}")
	{{- else }}
	if m.Has("{{ .ProtoName }}") {
		cols = append(cols, "{{ .SQLName }}")
	}
	{{- end }}
	{{- end }}
	for _, e := range extra {
		if !{{ $q }}.Contains(cols, e) {
			cols = append(cols, e)
		}
	}
	return cols
}

// Load{{ $msg }}Rows selects {{ .TableName }} rows matching conds and eagerly
// loads every relation the mask selects. extra forces columns into the SELECT
// regardless of the mask, which is how a parent gets the join key back.
func Load{{ $msg }}Rows(ctx {{ .CtxPkg }}.Context, c {{ $q }}.Conn, m *{{ $q }}.Mask, conds []{{ $q }}.Cond, extra ...string) ([]*{{ $msg }}Result, error) {
	cols := {{ $msg }}QueryColumns(m, extra...)
	sqlRows, err := c.Select(ctx, "{{ .TableName }}", cols, conds)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	var rows []*{{ $msg }}Result
	for sqlRows.Next() {
		row := new({{ $msg }}Result)
		if err := row.Scan(cols, sqlRows); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	if err := sqlRows.Err(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return rows, nil
	}
{{ range .Relations }}
	if m.HasRelation("{{ .ProtoName }}") {
		byKey := make(map[any][]*{{ $msg }}Result, len(rows))
		for _, row := range rows {
			key := {{ $q }}.Key({{ .OwnerKeyExpr }})
			if key == nil {
				continue
			}
			byKey[key] = append(byKey[key], row)
		}
		if cond, ok := {{ $q }}.In("{{ .RemoteSQLCol }}", {{ $q }}.Keys(byKey)); ok {
			related, err := {{ .LoaderFunc }}(ctx, c, m.Sub("{{ .ProtoName }}"), []{{ $q }}.Cond{cond}{{ if .IsList }}, "{{ .RemoteSQLCol }}"{{ end }})
			if err != nil {
				return nil, err
			}
			for _, rel := range related {
				for _, row := range byKey[{{ $q }}.Key({{ .RelatedKeyExpr }})] {
					{{ if .IsList }}row.{{ .GoName }} = append(row.{{ .GoName }}, &rel.{{ .TargetName }}){{ else }}row.{{ .GoName }} = &rel.{{ .TargetName }}{{ end }}
				}
			}
		}
	}
{{ end }}
	return rows, nil
}

// Load{{ $msg }} selects {{ .TableName }} rows restricted to the fields mask
// names. A nil or empty mask selects every column and loads no relations.
func Load{{ $msg }}(ctx {{ .CtxPkg }}.Context, c {{ $q }}.Conn, mask *{{ .FMPkg }}.FieldMask, conds ...{{ $q }}.Cond) ([]*{{ $msg }}, error) {
	rows, err := Load{{ $msg }}Rows(ctx, c, {{ $q }}.FromFieldMask(mask), conds)
	if err != nil {
		return nil, err
	}
	out := make([]*{{ $msg }}, len(rows))
	for i, row := range rows {
		out[i] = &row.{{ $msg }}
	}
	return out, nil
}
`

type Column struct {
	ProtoName string
	SQLName   string
	IsPK      bool
}

type Relation struct {
	ProtoName string
	GoName    string
	IsList    bool
	// TargetName is the embedded message field on the related Result struct.
	TargetName string
	// LoaderFunc is Load<Target>Rows, qualified if the target is in another package.
	LoaderFunc string
	// RemoteSQLCol is the target column used in the IN clause.
	RemoteSQLCol string
	// OwnerKeyExpr and RelatedKeyExpr read the join key off each side.
	OwnerKeyExpr   string
	RelatedKeyExpr string
}

type Table struct {
	Name      string
	TableName string
	QueryPkg  string
	FMPkg     string
	CtxPkg    string
	Columns   []*Column
	Relations []*Relation
}

var tpl = template.Must(template.New("query").Parse(templateStr))

type QueryWriter struct {
	o    writer.Printer
	repo types.TableRepo
}

func New(g writer.Printer, repo types.TableRepo) writer.Writer {
	return &QueryWriter{o: g, repo: repo}
}

func (s *QueryWriter) Write(protoFile *protogen.File) error {
	tables := s.repo.ForFile(protoFile)
	if len(tables) == 0 {
		return nil
	}
	for _, table := range tables {
		if err := s.table(table); err != nil {
			return err
		}
	}
	return nil
}

func (s *QueryWriter) table(table *types.Table) error {
	tbl := Table{
		Name:      table.GetMessageName(),
		TableName: table.GetTableName(),
		QueryPkg:  s.pkgAlias("Conn", queryImportPath, ".Conn"),
		FMPkg:     s.pkgAlias("FieldMask", fieldmaskImportPath, ".FieldMask"),
		CtxPkg:    s.pkgAlias("Context", contextImportPath, ".Context"),
	}

	for _, c := range table.GetColumns() {
		tbl.Columns = append(tbl.Columns, &Column{
			ProtoName: c.GetProtoName(),
			SQLName:   c.GetFieldname(),
			IsPK:      c.IsPrimaryKey(),
		})
	}

	for _, rel := range table.GetRelations() {
		out, err := s.relation(table, rel)
		if err != nil {
			return err
		}
		tbl.Relations = append(tbl.Relations, out)
	}

	return tpl.Execute(s.o, tbl)
}

// relation resolves one relation into the expressions the template needs. The
// directions differ only in which side holds the join key.
func (s *QueryWriter) relation(table *types.Table, rel *types.Relation) (*Relation, error) {
	target, err := rel.GetTarget(s.repo)
	if err != nil {
		return nil, err
	}
	remoteCols, err := rel.GetTargetColumns(s.repo)
	if err != nil {
		return nil, err
	}
	localCols, err := rel.GetLocalColumns(s.repo)
	if err != nil {
		return nil, err
	}
	if len(remoteCols) != 1 || len(localCols) != 1 {
		return nil, fmt.Errorf("table %q relation %q: composite keys are not supported yet (%d local, %d remote columns)",
			table.GetTableName(), rel.GetName(), len(localCols), len(remoteCols))
	}
	remote, local := remoteCols[0], localCols[0]

	loader := fmt.Sprintf("Load%sRows", target.GetMessageName())
	if target.File.GoImportPath != table.File.GoImportPath {
		loader = s.o.QualifiedGoIdent(protogen.GoIdent{
			GoName:       loader,
			GoImportPath: target.File.GoImportPath,
		})
	}

	return &Relation{
		ProtoName:      rel.GetProtoName(),
		GoName:         rel.GetName(),
		IsList:         rel.IsList(),
		TargetName:     target.GetMessageName(),
		LoaderFunc:     loader,
		RemoteSQLCol:   remote.GetFieldname(),
		OwnerKeyExpr:   keyExpr("row", local),
		RelatedKeyExpr: keyExpr("rel", remote),
	}, nil
}

// keyExpr reads a join key off a Result. A message-kind column has no usable
// getter, so the raw value the scanner stashed in fk_<column>_id is used.
func keyExpr(recv string, col *types.Column) string {
	if col.IsMessage() {
		return fmt.Sprintf("%s.fk_%s_id", recv, col.GetFieldname())
	}
	return fmt.Sprintf("%s.Get%s()", recv, col.GetName())
}

// pkgAlias resolves the local alias protogen assigned to an import.
func (s *QueryWriter) pkgAlias(goName string, path protogen.GoImportPath, suffix string) string {
	return strings.TrimSuffix(s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       goName,
		GoImportPath: path,
	}), suffix)
}
