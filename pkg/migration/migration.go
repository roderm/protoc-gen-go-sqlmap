// Package migration drives ariga/atlas migrations directly, from the
// resolved schema (*schemav1.SchemaTable) that protoc-gen-go-sqlmap emits
// into generated .sqlmap.go files.
package migration

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/mysql"
	"ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlite"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
)

// Dialect names accepted by New.
const (
	DialectPostgres = "postgres"
	DialectMySQL    = "mysql"
	DialectSQLite   = "sqlite3"
)

// Migrator drives schema migrations for a database/sql connection.
type Migrator struct {
	drv     migrate.Driver
	dialect string
	dir     migrate.Dir
	format  migrate.Formatter
}

type options struct {
	dir    migrate.Dir
	format migrate.Formatter
}

// Option configures a Migrator.
type Option func(*options)

// WithDir sets the atlas migration directory used by Diff and ApplyPending.
func WithDir(dir migrate.Dir) Option {
	return func(o *options) { o.dir = dir }
}

// WithFormatter sets the formatter used to write new migration files.
// Defaults to atlas's own format; see ariga.io/atlas/sql/sqltool for
// golang-migrate/goose/flyway/liquibase/dbmate formatters.
func WithFormatter(f migrate.Formatter) Option {
	return func(o *options) { o.format = f }
}

// New creates a Migrator for the given database/sql connection and dialect
// name (DialectPostgres, DialectMySQL, or DialectSQLite).
func New(db *sql.DB, dialect string, opts ...Option) (*Migrator, error) {
	drv, err := openDriver(dialect, db)
	if err != nil {
		return nil, err
	}
	cfg := &options{format: migrate.DefaultFormatter}
	for _, opt := range opts {
		opt(cfg)
	}
	return &Migrator{drv: drv, dialect: dialect, dir: cfg.dir, format: cfg.format}, nil
}

func openDriver(dialect string, db *sql.DB) (migrate.Driver, error) {
	switch dialect {
	case DialectPostgres:
		return postgres.Open(db)
	case DialectMySQL:
		return mysql.Open(db)
	case DialectSQLite:
		return sqlite.Open(db)
	default:
		return nil, fmt.Errorf("migration: unsupported dialect %q", dialect)
	}
}

// Create directly creates/alters tables in the database to match the given
// schema, declaratively — no migration history, no files.
func (m *Migrator) Create(ctx context.Context, tables ...*schemav1.SchemaTable) error {
	current, desired, err := m.diffState(ctx, tables)
	if err != nil {
		return err
	}
	changes, err := m.drv.SchemaDiff(current, desired)
	if err != nil {
		return fmt.Errorf("migration: diff: %w", err)
	}
	if err := m.drv.ApplyChanges(ctx, changes); err != nil {
		return fmt.Errorf("migration: apply: %w", err)
	}
	return nil
}

// Diff compares the current database state with the desired schema and
// writes a new migration file into the configured Dir (see WithDir).
func (m *Migrator) Diff(ctx context.Context, name string, tables ...*schemav1.SchemaTable) error {
	if m.dir == nil {
		return fmt.Errorf("migration: no migration directory configured, use WithDir")
	}
	current, desired, err := m.diffState(ctx, tables)
	if err != nil {
		return err
	}
	changes, err := m.drv.SchemaDiff(current, desired)
	if err != nil {
		return fmt.Errorf("migration: diff: %w", err)
	}
	plan, err := m.drv.PlanChanges(ctx, name, changes)
	if err != nil {
		return fmt.Errorf("migration: plan: %w", err)
	}
	files, err := m.format.Format(plan)
	if err != nil {
		return fmt.Errorf("migration: format: %w", err)
	}
	for _, f := range files {
		if err := m.dir.WriteFile(f.Name(), f.Bytes()); err != nil {
			return fmt.Errorf("migration: write %s: %w", f.Name(), err)
		}
	}
	return nil
}

// ApplyPending executes up to n pending migration files from the configured
// Dir against the connected database. n<=0 applies all pending files.
func (m *Migrator) ApplyPending(ctx context.Context, n int) error {
	if m.dir == nil {
		return fmt.Errorf("migration: no migration directory configured, use WithDir")
	}
	ex, err := migrate.NewExecutor(m.drv, m.dir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return fmt.Errorf("migration: %w", err)
	}
	return ex.ExecuteN(ctx, n)
}

// diffState scopes the desired schema to the inspected one's name, so
// SchemaDiff compares table contents rather than whole schemas.
func (m *Migrator) diffState(ctx context.Context, tables []*schemav1.SchemaTable) (current, desired *schema.Schema, err error) {
	realm, err := m.drv.InspectRealm(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("migration: inspect: %w", err)
	}
	if len(realm.Schemas) == 0 {
		return nil, nil, fmt.Errorf("migration: no schema found on the connected database")
	}
	current = realm.Schemas[0]
	desired, err = toSchema(m.dialect, current.Name, tables)
	if err != nil {
		return nil, nil, err
	}
	return current, desired, nil
}

// toSchema converts the resolved schema into atlas's types. One call must
// include every table referenced by a foreign key among them: resolution goes
// by pointer identity of the package-level vars the generator emits.
func toSchema(dialect, name string, tables []*schemav1.SchemaTable) (*schema.Schema, error) {
	tblMap := make(map[*schemav1.SchemaTable]*schema.Table, len(tables))
	colMap := make(map[*schemav1.SchemaColumn]*schema.Column)
	sch := schema.New(name)

	for _, t := range tables {
		at := schema.NewTable(t.GetName())
		var pk []*schema.Column
		for _, c := range t.GetColumns() {
			typ, err := parseType(dialect, c.GetType()[dialect])
			if err != nil {
				return nil, fmt.Errorf("migration: table %q column %q: %w", t.GetName(), c.GetName(), err)
			}
			ac := schema.NewColumn(c.GetName()).SetType(typ).SetNull(c.GetNullable())
			if expr := c.GetDefaultExpr(); expr != "" {
				ac.SetDefault(&schema.RawExpr{X: expr})
			}
			if c.GetPk() == schemav1.PK_PK_AUTO {
				if err := addAutoIncrement(dialect, ac); err != nil {
					return nil, fmt.Errorf("migration: table %q column %q: %w", t.GetName(), c.GetName(), err)
				}
			}
			at.AddColumns(ac)
			if c.GetPk() != schemav1.PK_PK_UNSPECIFIED {
				pk = append(pk, ac)
			}
			colMap[c] = ac
		}
		if len(pk) > 0 {
			at.SetPrimaryKey(schema.NewPrimaryKey(pk...))
		}
		for _, chk := range t.GetChecks() {
			expr, ok := chk.GetExpr()[dialect]
			if !ok {
				return nil, fmt.Errorf("migration: table %q check %q has no expression for dialect %q", t.GetName(), chk.GetName(), dialect)
			}
			at.AddChecks(schema.NewCheck().SetName(chk.GetName()).SetExpr(expr))
		}
		for _, idx := range t.GetIndexes() {
			cols, err := lookupColumns(colMap, idx.GetColumns())
			if err != nil {
				return nil, fmt.Errorf("migration: table %q index %q: %w", t.GetName(), idx.GetName(), err)
			}
			ai := schema.NewIndex(idx.GetName()).SetUnique(idx.GetUnique()).AddColumns(cols...)
			at.AddIndexes(ai)
		}
		tblMap[t] = at
		sch.AddTables(at)
	}

	for _, t := range tables {
		at := tblMap[t]
		for _, fk := range t.GetForeignKeys() {
			refTbl, ok := tblMap[fk.GetRefTable()]
			if !ok {
				return nil, fmt.Errorf("migration: table %q has a foreign key referencing table %q, which was not included in this call", t.GetName(), fk.GetRefTable().GetName())
			}
			cols, err := lookupColumns(colMap, fk.GetColumns())
			if err != nil {
				return nil, fmt.Errorf("migration: table %q: %w", t.GetName(), err)
			}
			refCols, err := lookupColumns(colMap, fk.GetRefColumns())
			if err != nil {
				return nil, fmt.Errorf("migration: table %q: %w", t.GetName(), err)
			}
			if fk.GetOnDelete() == schemav1.OnDelete_ON_DELETE_SET_NULL {
				for _, c := range cols {
					if !c.Type.Null {
						return nil, fmt.Errorf("migration: table %q: foreign key column %q is NOT NULL but ON DELETE SET NULL was requested", t.GetName(), c.Name)
					}
				}
			}
			afk := schema.NewForeignKey(fmt.Sprintf("%s_%s_fk", t.GetName(), refTbl.Name)).
				AddColumns(cols...).
				SetRefTable(refTbl).
				AddRefColumns(refCols...).
				SetOnDelete(toReferenceOption(fk.GetOnDelete()))
			at.AddForeignKeys(afk)
		}
	}
	return sch, nil
}

// parseType parses a raw SQL type string into atlas's schema.Type.
func parseType(dialect, raw string) (schema.Type, error) {
	switch dialect {
	case DialectPostgres:
		return postgres.ParseType(strings.ToLower(raw))
	case DialectMySQL:
		return mysql.ParseType(strings.ToLower(raw))
	case DialectSQLite:
		return sqlite.ParseType(strings.ToLower(raw))
	default:
		return nil, fmt.Errorf("unsupported dialect %q", dialect)
	}
}

// addAutoIncrement attaches the dialect-specific auto-generate attribute;
// atlas has no generic concept for it.
func addAutoIncrement(dialect string, col *schema.Column) error {
	// Every dialect restricts generated columns to integers. Catching it here
	// names the column instead of failing opaquely in the database.
	if _, ok := col.Type.Type.(*schema.IntegerType); !ok {
		return fmt.Errorf("PK_AUTO requires an integer column, but %q is %T; use PK_MAN for an application-supplied key", col.Name, col.Type.Type)
	}
	switch dialect {
	case DialectMySQL:
		col.AddAttrs(&mysql.AutoIncrement{})
	case DialectSQLite:
		col.AddAttrs(&sqlite.AutoIncrement{})
	case DialectPostgres:
		col.AddAttrs(&postgres.Identity{Generation: "BY DEFAULT"})
	default:
		return fmt.Errorf("unsupported dialect %q", dialect)
	}
	return nil
}

func lookupColumns(colMap map[*schemav1.SchemaColumn]*schema.Column, cols []*schemav1.SchemaColumn) ([]*schema.Column, error) {
	out := make([]*schema.Column, len(cols))
	for i, c := range cols {
		ac, ok := colMap[c]
		if !ok {
			return nil, fmt.Errorf("column %q was not included in this call", c.GetName())
		}
		out[i] = ac
	}
	return out, nil
}

func toReferenceOption(od schemav1.OnDelete) schema.ReferenceOption {
	switch od {
	case schemav1.OnDelete_ON_DELETE_CASCADE:
		return schema.Cascade
	case schemav1.OnDelete_ON_DELETE_SET_NULL:
		return schema.SetNull
	default:
		return schema.NoAction
	}
}
