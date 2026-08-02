package migration

import (
	"testing"

	"ariga.io/atlas/sql/schema"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
)

func newCol(name string, pk schemav1.PK) *schemav1.SchemaColumn {
	return &schemav1.SchemaColumn{
		Name: &name,
		Type: map[string]string{DialectPostgres: "VARCHAR(255)"},
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

func strPtr(s string) *string { return &s }
