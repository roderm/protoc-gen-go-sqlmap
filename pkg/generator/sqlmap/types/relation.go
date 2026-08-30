package types

import (
	"fmt"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"google.golang.org/protobuf/compiler/protogen"
)

// Relation is a message-kind field carrying a foreign_key. Its cardinality
// picks the direction: singular is belongs-to (this row holds the key),
// repeated is has-many (the target rows do). ForeignKey.Fieldnames always
// names columns on the target table.
type Relation struct {
	Table *Table
	Def   *sqlmapv1.Column
	Field *protogen.Field
}

// IsList reports whether this is a has-many relation.
func (r *Relation) IsList() bool { return r.Field.Desc.IsList() }

// GetName returns the Go field name holding the relation, e.g. "Books".
func (r *Relation) GetName() string { return r.Field.GoName }

// GetProtoName returns the proto field name, as used in a FieldMask path.
func (r *Relation) GetProtoName() string { return string(r.Field.Desc.Name()) }

// GetTarget resolves the table this relation points at.
func (r *Relation) GetTarget(repo TableRepo) (*Table, error) {
	entity := string(r.Field.Message.Desc.Name())
	if r.Def.ForeignKey.Entity != nil {
		entity = r.Def.ForeignKey.GetEntity()
	}
	tbl, ok := repo.GetByName(entity)
	if !ok {
		return nil, fmt.Errorf("relation %q references unknown entity %q", r.GetName(), entity)
	}
	return tbl, nil
}

// GetTargetColumns returns the target's key columns, defaulting to its PK.
func (r *Relation) GetTargetColumns(repo TableRepo) ([]*Column, error) {
	target, err := r.GetTarget(repo)
	if err != nil {
		return nil, err
	}
	cols, err := ResolveRefColumns(target, r.Def.ForeignKey.GetFieldnames())
	if err != nil {
		return nil, fmt.Errorf("relation %q: %w", r.GetName(), err)
	}
	return cols, nil
}

// ResolveRefColumns resolves a foreign key's `fieldnames` against the target,
// defaulting to its primary key. Names match the SQL column name first and the
// Go field name second, since both spellings appear in the wild.
func ResolveRefColumns(target *Table, names []string) ([]*Column, error) {
	if len(names) == 0 {
		pks := target.GetPKColumns()
		if len(pks) == 0 {
			return nil, fmt.Errorf("target %q has no primary key to reference", target.GetMessageName())
		}
		return pks, nil
	}
	cols := make([]*Column, 0, len(names))
	for _, name := range names {
		col, ok := target.GetColumnByFieldName(name)
		if !ok {
			col, ok = target.GetColumnByName(name)
		}
		if !ok {
			return nil, fmt.Errorf("target %q has no column %q", target.GetMessageName(), name)
		}
		cols = append(cols, col)
	}
	return cols, nil
}

// GetLocalColumns returns this table's join columns: the PK for a has-many,
// the field's own column for a belongs-to.
func (r *Relation) GetLocalColumns(repo TableRepo) ([]*Column, error) {
	if r.IsList() {
		pks := r.Table.GetPKColumns()
		if len(pks) == 0 {
			return nil, fmt.Errorf("relation %q: %q has no primary key for its children to reference", r.GetName(), r.Table.GetMessageName())
		}
		return pks, nil
	}
	col, ok := r.Table.GetColumnByName(r.GetName())
	if !ok {
		return nil, fmt.Errorf("relation %q: no column found on %q", r.GetName(), r.Table.GetMessageName())
	}
	return []*Column{col}, nil
}

// GetPKColumns returns every column marked as a primary key, in field order.
func (t *Table) GetPKColumns() []*Column {
	var pks []*Column
	for _, col := range t.GetColumns() {
		if col.Def.GetPk() != schemav1.PK_PK_UNSPECIFIED {
			pks = append(pks, col)
		}
	}
	return pks
}

// GetColumnByName looks a column up by its Go field name.
func (t *Table) GetColumnByName(name string) (*Column, bool) {
	for _, col := range t.GetColumns() {
		if col.GetName() == name {
			return col, true
		}
	}
	return nil, false
}
