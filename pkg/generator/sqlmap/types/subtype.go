package types

import (
	"fmt"

	sqlmapv1 "github.com/snaerverk/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// defaultDiscriminatorType holds a proto field name comfortably.
const defaultDiscriminatorType = "VARCHAR(32)"

// Subtype is one arm of a supertype's oneof.
type Subtype struct {
	Field *protogen.Field // oneof member on the supertype
	Value string          // what the discriminator holds for it
	Table *Table
}

// Hierarchy is a supertype table together with the subtypes its oneof names.
type Hierarchy struct {
	Super    *Table // carries the discriminator column
	Oneof    *protogen.Oneof
	Def      *sqlmapv1.Subtypes
	Subtypes []*Subtype // arms, in declaration order
}

// GetDiscriminatorName returns the discriminator column's SQL name.
func (h *Hierarchy) GetDiscriminatorName() string { return h.Def.GetDiscriminator() }

// GetDiscriminatorType defaults to VARCHAR(32).
func (h *Hierarchy) GetDiscriminatorType(dialect string) string {
	if t, ok := h.Def.GetType()[dialect]; ok {
		return t
	}
	return defaultDiscriminatorType
}

// GetHierarchy returns the subtype hierarchy rooted at this table, or nil when
// no oneof carries the `subtypes` option.
func (t *Table) GetHierarchy(repo TableRepo) (*Hierarchy, error) {
	for _, oneof := range t.Msg.Oneofs {
		ext := proto.GetExtension(oneof.Desc.Options(), sqlmapv1.E_Subtypes).(*sqlmapv1.Subtypes)
		if ext == nil {
			continue
		}
		if ext.GetDiscriminator() == "" {
			return nil, fmt.Errorf("table %q: oneof %q declares subtypes with no discriminator column", t.GetTableName(), oneof.Desc.Name())
		}
		h := &Hierarchy{Super: t, Oneof: oneof, Def: ext}
		for _, f := range oneof.Fields {
			if f.Desc.Kind() != protoreflect.MessageKind {
				return nil, fmt.Errorf("table %q: subtype oneof member %q is not a message", t.GetTableName(), f.Desc.Name())
			}
			sub, ok := repo.GetByName(string(f.Message.Desc.Name()))
			if !ok {
				return nil, fmt.Errorf("table %q: subtype %q is not a table", t.GetTableName(), f.Message.Desc.Name())
			}
			h.Subtypes = append(h.Subtypes, &Subtype{
				Field: f,
				Value: subtypeValue(sub, f),
				Table: sub,
			})
		}
		if len(h.Subtypes) == 0 {
			return nil, fmt.Errorf("table %q: subtype oneof %q has no members", t.GetTableName(), oneof.Desc.Name())
		}
		return h, nil
	}
	return nil, nil
}

// subtypeValue is the declared value, or the oneof field's name.
func subtypeValue(sub *Table, f *protogen.Field) string {
	if v := sub.Def.GetSubtypeOf().GetValue(); v != "" {
		return v
	}
	return string(f.Desc.Name())
}

// GetSuper resolves the supertype this table details and the oneof arm
// pointing back at it, or nil when it declares no `subtype_of`.
func (t *Table) GetSuper(repo TableRepo) (*Hierarchy, *Subtype, error) {
	def := t.Def.GetSubtypeOf()
	if def == nil {
		return nil, nil, nil
	}
	super, ok := repo.GetByName(def.GetEntity())
	if !ok {
		return nil, nil, fmt.Errorf("table %q: subtype_of names unknown entity %q", t.GetTableName(), def.GetEntity())
	}
	h, err := super.GetHierarchy(repo)
	if err != nil {
		return nil, nil, err
	}
	if h == nil {
		return nil, nil, fmt.Errorf("table %q is a subtype of %q, but %q has no oneof carrying the subtypes option",
			t.GetTableName(), super.GetMessageName(), super.GetMessageName())
	}
	for _, sub := range h.Subtypes {
		if sub.Table == t {
			return h, sub, nil
		}
	}
	return nil, nil, fmt.Errorf("table %q is a subtype of %q, but %q's subtype oneof does not name it",
		t.GetTableName(), super.GetMessageName(), super.GetMessageName())
}

// GetSubtypeKeyColumns returns this table's link columns, defaulting to its PK.
func (t *Table) GetSubtypeKeyColumns() ([]*Column, error) {
	names := t.Def.GetSubtypeOf().GetFieldnames()
	if len(names) == 0 {
		pks := t.GetPKColumns()
		if len(pks) == 0 {
			return nil, fmt.Errorf("table %q is a subtype but has no primary key to link on", t.GetTableName())
		}
		return pks, nil
	}
	cols := make([]*Column, 0, len(names))
	for _, name := range names {
		col, ok := t.GetColumnByFieldName(name)
		if !ok {
			col, ok = t.GetColumnByName(name)
		}
		if !ok {
			return nil, fmt.Errorf("table %q: subtype_of names unknown column %q", t.GetTableName(), name)
		}
		cols = append(cols, col)
	}
	return cols, nil
}
