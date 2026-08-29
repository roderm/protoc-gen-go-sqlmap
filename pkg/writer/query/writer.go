// Package query is the writer that emits FieldMask-aware row loading with
// eager loading of related rows.
//
// Related rows are fetched with one batched `IN (...)` query per relation and
// stitched together in Go, rather than joined -- a join would multiply the
// parent row out once per child and force the scanner to de-duplicate it. The
// FieldMask drives both halves: which columns land in the SELECT, and which
// relations are loaded at all, so an unmasked relation costs no query.
package query

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"google.golang.org/protobuf/compiler/protogen"
)

const (
	queryImportPath     = protogen.GoImportPath("github.com/roderm/protoc-gen-go-sqlmap/pkg/query")
	fieldmaskImportPath = protogen.GoImportPath("google.golang.org/protobuf/types/known/fieldmaskpb")
	contextImportPath   = protogen.GoImportPath("context")
)

const templateStr = `
{{- $q := .QueryPkg }}
{{- $msg := .Name }}
// {{ $msg }}QueryColumns returns the columns to select for mask. Primary keys
// are always included: they are what relations are stitched on, so leaving
// them out of the mask must not break eager loading.
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
// loads every relation the mask selects. extra names columns that must be in
// the SELECT regardless of the mask, which is how a parent load makes sure the
// join key comes back on its children.
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
	if m.Has("{{ .ProtoName }}") {
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

// Load{{ $msg }} selects {{ .TableName }} rows, restricted to the fields named
// by mask. A nil or empty mask selects every column and no relations.
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
	// LoaderFunc is Load<Target>Rows, import-qualified when the target lives
	// in another Go package.
	LoaderFunc string
	// RemoteSQLCol is the column on the target table used in the IN clause.
	RemoteSQLCol string
	// OwnerKeyExpr reads the join key off a row of *this* table; for a
	// belongs-to that is the raw foreign-key value the scanner stashed.
	OwnerKeyExpr string
	// RelatedKeyExpr reads the same key off a related row.
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

// relation resolves one relation into the expressions the template needs.
//
// The two directions differ only in where the join key lives. For a has-many
// the owner supplies its primary key and the related rows carry the foreign
// key; for a belongs-to the owner carries the foreign key and the related rows
// are matched on the column it points at.
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

// keyExpr renders the expression reading a join key off a Result value. A
// message-kind column has no usable Go getter -- the message field is what the
// relation will eventually be filled with -- so the raw value the scanner
// stashed in fk_<column>_id is used instead.
func keyExpr(recv string, col *types.Column) string {
	if col.IsMessage() {
		return fmt.Sprintf("%s.fk_%s_id", recv, col.GetFieldname())
	}
	return fmt.Sprintf("%s.Get%s()", recv, col.GetName())
}

// pkgAlias resolves the local alias protogen assigned to an import, by asking
// for a known identifier from it and trimming that identifier back off.
func (s *QueryWriter) pkgAlias(goName string, path protogen.GoImportPath, suffix string) string {
	return strings.TrimSuffix(s.o.QualifiedGoIdent(protogen.GoIdent{
		GoName:       goName,
		GoImportPath: path,
	}), suffix)
}
