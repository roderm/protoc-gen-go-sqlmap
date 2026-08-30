package migration

import (
	"strings"
	"testing"

	"ariga.io/atlas/sql/schema"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
)

// newCol builds a VARCHAR test column, except for PK_AUTO columns: an
// auto-generated key has to be an integer in every dialect.
func newCol(name string, pk schemav1.PK) *schemav1.SchemaColumn {
	typ := "VARCHAR(255)"
	if pk == schemav1.PK_PK_AUTO {
		typ = "BIGINT"
	}
	return &schemav1.SchemaColumn{
		Name: &name,
		Type: map[string]string{DialectPostgres: typ},
		Pk:   pk.Enum(),
	}
}

func TestToSchema_ForeignKeyPointerIdentity(t *testing.T) {
	parentID := newCol("parent_id", schemav1.PK_PK_AUTO)
	parent := &schemav1.SchemaTable{
		Name:    strPtr("tbl_parent"),
		Columns: []*schemav1.SchemaColumn{parentID},
	}

	childID := newCol("child_id", schemav1.PK_PK_AUTO)
	childParentID := newCol("parent_id", schemav1.PK_PK_UNSPECIFIED)
	child := &schemav1.SchemaTable{
		Name:    strPtr("tbl_child"),
		Columns: []*schemav1.SchemaColumn{childID, childParentID},
		ForeignKeys: []*schemav1.SchemaForeignKey{
			{
				Columns:    []*schemav1.SchemaColumn{childParentID},
				RefTable:   parent,
				RefColumns: []*schemav1.SchemaColumn{parentID},
				OnDelete:   schemav1.OnDelete_ON_DELETE_CASCADE.Enum(),
			},
		},
	}

	sch, err := toSchema(DialectPostgres, "public", []*schemav1.SchemaTable{parent, child})
	if err != nil {
		t.Fatalf("toSchema: %v", err)
	}

	at, ok := sch.Table("tbl_child")
	if !ok {
		t.Fatalf("tbl_child not found in converted schema")
	}
	if len(at.ForeignKeys) != 1 {
		t.Fatalf("expected 1 foreign key, got %d", len(at.ForeignKeys))
	}
	fk := at.ForeignKeys[0]

	pt, ok := sch.Table("tbl_parent")
	if !ok {
		t.Fatalf("tbl_parent not found in converted schema")
	}

	// The FK's Columns/RefColumns must be the exact same *schema.Column
	// instances already attached to their owning tables (pointer identity),
	// not just same-named copies -- that's what lets atlas resolve the FK.
	childCol, ok := at.Column("parent_id")
	if !ok || fk.Columns[0] != childCol {
		t.Errorf("fk.Columns[0] is not the same *schema.Column instance as tbl_child.parent_id")
	}
	parentCol, ok := pt.Column("parent_id")
	if !ok || fk.RefColumns[0] != parentCol {
		t.Errorf("fk.RefColumns[0] is not the same *schema.Column instance as tbl_parent.parent_id")
	}
	if fk.RefTable != pt {
		t.Errorf("fk.RefTable is not the same *schema.Table instance as tbl_parent")
	}
	if fk.OnDelete != schema.Cascade {
		t.Errorf("fk.OnDelete = %q, want %q", fk.OnDelete, schema.Cascade)
	}
}

func TestToSchema_AutoIncrementPerDialect(t *testing.T) {
	id := newCol("id", schemav1.PK_PK_AUTO)
	id.Type = map[string]string{
		DialectPostgres: "BIGINT",
		DialectMySQL:    "BIGINT",
		DialectSQLite:   "INTEGER",
	}
	tbl := &schemav1.SchemaTable{Name: strPtr("tbl"), Columns: []*schemav1.SchemaColumn{id}}

	for _, dialect := range []string{DialectPostgres, DialectMySQL, DialectSQLite} {
		sch, err := toSchema(dialect, "public", []*schemav1.SchemaTable{tbl})
		if err != nil {
			t.Fatalf("toSchema(%s): %v", dialect, err)
		}
		at, _ := sch.Table("tbl")
		col, ok := at.Column("id")
		if !ok {
			t.Fatalf("%s: column id not found", dialect)
		}
		if len(col.Attrs) == 0 {
			t.Errorf("%s: expected an auto-increment attribute on PK_AUTO column, got none", dialect)
		}
	}
}

// PK_AUTO on a VARCHAR is an easy mistake when the id is really an
// application-supplied UUID. Every dialect rejects a generated non-integer
// column, so it has to fail here rather than as opaque DDL from the database.
func TestToSchema_AutoIncrementRequiresInteger(t *testing.T) {
	id := newCol("id", schemav1.PK_PK_AUTO)
	id.Type = map[string]string{DialectPostgres: "VARCHAR(64)"}
	tbl := &schemav1.SchemaTable{Name: strPtr("tbl"), Columns: []*schemav1.SchemaColumn{id}}

	_, err := toSchema(DialectPostgres, "public", []*schemav1.SchemaTable{tbl})
	if err == nil {
		t.Fatal("expected an error for PK_AUTO on a non-integer column, got nil")
	}
	if !strings.Contains(err.Error(), "PK_MAN") {
		t.Errorf("error should point at PK_MAN as the fix, got: %v", err)
	}
}

func TestToSchema_MissingReferencedTable(t *testing.T) {
	otherID := newCol("id", schemav1.PK_PK_AUTO)
	other := &schemav1.SchemaTable{Name: strPtr("tbl_other"), Columns: []*schemav1.SchemaColumn{otherID}}

	fkCol := newCol("other_id", schemav1.PK_PK_UNSPECIFIED)
	tbl := &schemav1.SchemaTable{
		Name:    strPtr("tbl"),
		Columns: []*schemav1.SchemaColumn{fkCol},
		ForeignKeys: []*schemav1.SchemaForeignKey{{
			Columns:    []*schemav1.SchemaColumn{fkCol},
			RefTable:   other,
			RefColumns: []*schemav1.SchemaColumn{otherID},
		}},
	}

	// other is deliberately NOT included in this call.
	if _, err := toSchema(DialectPostgres, "public", []*schemav1.SchemaTable{tbl}); err == nil {
		t.Fatal("expected an error when a foreign key references a table not included in the call, got nil")
	}
}

func TestToSchema_Nullability(t *testing.T) {
	id := newCol("id", schemav1.PK_PK_AUTO)
	nullable := newCol("nick", schemav1.PK_PK_UNSPECIFIED)
	nullable.Nullable = boolPtr(true)
	notNull := newCol("name", schemav1.PK_PK_UNSPECIFIED)

	tbl := &schemav1.SchemaTable{
		Name:    strPtr("tbl"),
		Columns: []*schemav1.SchemaColumn{id, nullable, notNull},
	}
	sch, err := toSchema(DialectPostgres, "public", []*schemav1.SchemaTable{tbl})
	if err != nil {
		t.Fatalf("toSchema: %v", err)
	}
	at, _ := sch.Table("tbl")

	for name, want := range map[string]bool{"id": false, "nick": true, "name": false} {
		col, ok := at.Column(name)
		if !ok {
			t.Fatalf("column %q not found", name)
		}
		if col.Type.Null != want {
			t.Errorf("column %q Null = %v, want %v", name, col.Type.Null, want)
		}
	}
}

func TestToSchema_SetNullOnNotNullColumnRejected(t *testing.T) {
	parentID := newCol("parent_id", schemav1.PK_PK_AUTO)
	parent := &schemav1.SchemaTable{
		Name:    strPtr("tbl_parent"),
		Columns: []*schemav1.SchemaColumn{parentID},
	}

	// Deliberately NOT nullable, while the FK asks for ON DELETE SET NULL.
	childParentID := newCol("parent_id", schemav1.PK_PK_UNSPECIFIED)
	child := &schemav1.SchemaTable{
		Name:    strPtr("tbl_child"),
		Columns: []*schemav1.SchemaColumn{childParentID},
		ForeignKeys: []*schemav1.SchemaForeignKey{{
			Columns:    []*schemav1.SchemaColumn{childParentID},
			RefTable:   parent,
			RefColumns: []*schemav1.SchemaColumn{parentID},
			OnDelete:   schemav1.OnDelete_ON_DELETE_SET_NULL.Enum(),
		}},
	}

	if _, err := toSchema(DialectPostgres, "public", []*schemav1.SchemaTable{parent, child}); err == nil {
		t.Fatal("expected an error for ON DELETE SET NULL on a NOT NULL column, got nil")
	}
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
