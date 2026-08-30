package schema

// text/template, not html/template: HTML escaping mangles quotes in Go source.
import (
	"fmt"
	"strconv"
	"strings"
	"text/template"

	schemav1 "github.com/snaerverk/protoc-gen-go-sqlmap/pkg/generated/schema/v1"

	"github.com/snaerverk/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/snaerverk/protoc-gen-go-sqlmap/pkg/writer"
	"google.golang.org/protobuf/compiler/protogen"
)

// dialects every generated column carries a type for, so one file serves all.
var dialects = []string{"mysql", "postgres", "sqlite3"}

// schemaV1ImportPath holds the SchemaTable/SchemaColumn types this writer emits.
const schemaV1ImportPath = protogen.GoImportPath("github.com/snaerverk/protoc-gen-go-sqlmap/pkg/generated/schema/v1")

// protoImportPath supplies proto.String() for the schema.v1 *string fields.
const protoImportPath = protogen.GoImportPath("google.golang.org/protobuf/proto")

const templateStr = `
{{- $pkg := .SchemaPkg }}
{{- $proto := .ProtoPkg }}
var (
{{- $msg := .Name }}
{{- range .Columns }}
	{{ .VarName }} = &{{ $pkg }}.SchemaColumn{
		Name: {{ $proto }}.String("{{ .SQLName }}"),
		Type: map[string]string{
			{{ range $k, $v := .Type }}"{{ $k }}": "{{ $v }}",
			{{ end }}
		},
		Pk: {{ $pkg }}.{{ .Pk }}.Enum(),
		Nullable: {{ $proto }}.Bool({{ .Nullable }}),
		{{- if .DefaultExpr }}
		DefaultExpr: {{ $proto }}.String({{ q .DefaultExpr }}),
		{{- end }}
	}
{{- end }}
)
var {{ $msg }}Table = &{{ $pkg }}.SchemaTable{
	Name: {{ $proto }}.String("{{ .TableName }}"),
	Columns: []*{{ $pkg }}.SchemaColumn{
		{{- range .Columns }}
		{{ .VarName }},
		{{- end }}
	},
	{{- if .Checks }}
	Checks: []*{{ $pkg }}.SchemaCheck{
		{{- range .Checks }}
		{
			Name: {{ $proto }}.String("{{ .Name }}"),
			Expr: map[string]string{
				{{- range $k, $v := .Expr }}
				"{{ $k }}": {{ q $v }},
				{{- end }}
			},
		},
		{{- end }}
	},
	{{- end }}
	{{- if .Indexes }}
	Indexes: []*{{ $pkg }}.SchemaIndex{
		{{- range .Indexes }}
		{
			Name: {{ $proto }}.String("{{ .Name }}"),
			Unique: {{ $proto }}.Bool({{ .Unique }}),
			Columns: []*{{ $pkg }}.SchemaColumn{
				{{- range .Columns }}
				{{ . }},
				{{- end }}
			},
		},
		{{- end }}
	},
	{{- end }}
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
	VarName     string // Go var holding this column, e.g. "SimpleColumnDef_Id"
	SQLName     string // actual SQL column name, e.g. "simple_id"
	Type        map[string]string
	Pk          string
	Nullable    bool
	DefaultExpr string // raw SQL, emitted verbatim; empty means none
}

type Check struct {
	Name string
	// Expr is per dialect: identifier quoting differs, and getting it wrong on
	// MySQL yields a silently always-false constraint.
	Expr map[string]string
}

type Index struct {
	Name    string
	Columns []string // Go var names
	Unique  bool
}

type Table struct {
	Name        string
	TableName   string
	SchemaPkg   string
	ProtoPkg    string
	Columns     []*Column
	ForeignKeys []*ForeignKey
	Checks      []*Check
	Indexes     []*Index
}

type ForeignKey struct {
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
	OnDelete          string
}

// pkGoIdent returns the generated Go constant, which protoc-gen-go prefixes
// differently from the proto enum-value name.
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

// q renders a Go string literal. SQL fragments carry quotes of their own, and
// a backtick cannot appear in a Go raw string at all.
var tplFuncs = template.FuncMap{"q": strconv.Quote}

var tpl = template.Must(template.New("schema").Funcs(tplFuncs).Parse(templateStr))

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
			VarName:  columnVar(table, c.GetName()),
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
			out.Columns = append(out.Columns, columnVar(table, f.GetName()))
		}
		for _, name := range fk.To.GetFieldnames() {
			f, ok := refTable.GetColumnByFieldName(name)
			if !ok {
				return fmt.Errorf("table %q: referenced column %q not found on %q", table.GetTableName(), name, refTable.GetTableName())
			}
			out.ReferencedColumns = append(out.ReferencedColumns,
				s.colDefIdent(refTable, columnVar(refTable, f.GetName()), external))
		}
		tbl.ForeignKeys = append(tbl.ForeignKeys, out)
	}

	if err := s.subtypes(table, &tbl); err != nil {
		return err
	}

	return tpl.Execute(s.o, tbl)
}

// columnVar is the Go var holding a column definition.
func columnVar(table *types.Table, goName string) string {
	return fmt.Sprintf("%sColumnDef_%s", table.GetMessageName(), goName)
}

// discriminatorVar names the synthesized discriminator column. Deliberately
// unlike columnVar, which could collide with a real field.
func discriminatorVar(table *types.Table) string {
	return fmt.Sprintf("%sDiscriminatorColumn", table.GetMessageName())
}

// subtypes emits the joined-table subtype constraints, on either side. See
// docs/design/DESIGN-SUBTYPE-TABLES.md.
func (s *SchemaWriter) subtypes(table *types.Table, tbl *Table) error {
	if err := s.superTable(table, tbl); err != nil {
		return err
	}
	return s.subTable(table, tbl)
}

func (s *SchemaWriter) superTable(table *types.Table, tbl *Table) error {
	h, err := table.GetHierarchy(s.repo)
	if err != nil || h == nil {
		return err
	}
	pks := table.GetPKColumns()
	if len(pks) == 0 {
		return fmt.Errorf("table %q declares subtypes but has no primary key for them to reference", table.GetTableName())
	}

	name := h.GetDiscriminatorName()
	tbl.Columns = append(tbl.Columns, s.discriminatorColumn(table, h, ""))

	values := make([]string, len(h.Subtypes))
	for i, sub := range h.Subtypes {
		values[i] = sqlLiteral(sub.Value)
	}
	tbl.Checks = append(tbl.Checks, &Check{
		Name: fmt.Sprintf("%s_%s_check", table.GetTableName(), name),
		Expr: perDialect(func(d string) string {
			return fmt.Sprintf("%s IN (%s)", quoteIdent(d, name), strings.Join(values, ", "))
		}),
	})

	// A foreign key needs a uniquely-constrained target, even though the key
	// alone is already unique.
	idx := &Index{
		Name:   fmt.Sprintf("%s_%s_key", table.GetTableName(), name),
		Unique: true,
	}
	for _, pk := range pks {
		idx.Columns = append(idx.Columns, columnVar(table, pk.GetName()))
	}
	idx.Columns = append(idx.Columns, discriminatorVar(table))
	tbl.Indexes = append(tbl.Indexes, idx)
	return nil
}

func (s *SchemaWriter) subTable(table *types.Table, tbl *Table) error {
	h, sub, err := table.GetSuper(s.repo)
	if err != nil || h == nil {
		return err
	}
	keys, err := table.GetSubtypeKeyColumns()
	if err != nil {
		return err
	}
	super := h.Super
	superPKs := super.GetPKColumns()
	if len(keys) != len(superPKs) {
		return fmt.Errorf("table %q links to %q on %d column(s) but %q has %d primary key column(s)",
			table.GetTableName(), super.GetTableName(), len(keys), super.GetTableName(), len(superPKs))
	}

	name := h.GetDiscriminatorName()
	tbl.Columns = append(tbl.Columns, s.discriminatorColumn(table, h, sub.Value))
	tbl.Checks = append(tbl.Checks, &Check{
		Name: fmt.Sprintf("%s_%s_check", table.GetTableName(), name),
		Expr: perDialect(func(d string) string {
			return fmt.Sprintf("%s = %s", quoteIdent(d, name), sqlLiteral(sub.Value))
		}),
	})

	external := super.File.GoImportPath != table.File.GoImportPath
	fk := &ForeignKey{
		ReferencedTable: s.colDefIdent(super, fmt.Sprintf("%sTable", super.GetMessageName()), external),
		OnDelete:        onDeleteGoIdent(table.Def.GetSubtypeOf().GetOnDelete()),
	}
	for _, k := range keys {
		fk.Columns = append(fk.Columns, columnVar(table, k.GetName()))
	}
	fk.Columns = append(fk.Columns, discriminatorVar(table))
	for _, pk := range superPKs {
		fk.ReferencedColumns = append(fk.ReferencedColumns,
			s.colDefIdent(super, columnVar(super, pk.GetName()), external))
	}
	fk.ReferencedColumns = append(fk.ReferencedColumns,
		s.colDefIdent(super, discriminatorVar(super), external))
	tbl.ForeignKeys = append(tbl.ForeignKeys, fk)
	return nil
}

// discriminatorColumn builds the synthesized discriminator; a non-empty value
// pins it with a DEFAULT so subtype inserts need not name it.
func (s *SchemaWriter) discriminatorColumn(table *types.Table, h *types.Hierarchy, value string) *Column {
	col := &Column{
		VarName:  discriminatorVar(table),
		SQLName:  h.GetDiscriminatorName(),
		Pk:       pkGoIdent(schemav1.PK_PK_UNSPECIFIED),
		Nullable: false,
		Type:     make(map[string]string, len(dialects)),
	}
	for _, d := range dialects {
		col.Type[d] = h.GetDiscriminatorType(d)
	}
	if value != "" {
		col.DefaultExpr = sqlLiteral(value)
	}
	return col
}

// sqlLiteral renders a string as a SQL literal, doubling any embedded quote.
func sqlLiteral(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// perDialect builds the same expression for every supported dialect.
func perDialect(build func(dialect string) string) map[string]string {
	out := make(map[string]string, len(dialects))
	for _, d := range dialects {
		out[d] = build(d)
	}
	return out
}

// quoteIdent quotes an identifier for a CHECK expression. MySQL needs
// backticks: without ANSI_QUOTES it reads "kind" as a string literal, so the
// constraint is accepted and then rejects every row.
func quoteIdent(dialect, s string) string {
	if dialect == "mysql" {
		return "`" + strings.ReplaceAll(s, "`", "``") + "`"
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
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
