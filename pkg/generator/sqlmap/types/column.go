package types

import (
	"errors"
	"fmt"
	"slices"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Column struct {
	Table *Table
	Def   *sqlmapv1.Column
	Field *protogen.Field
}

func NewColumn(table *Table, field *protogen.Field) (*Column, error) {
	ext := proto.GetExtension(field.Desc.Options(), sqlmapv1.E_Col).(*sqlmapv1.Column)
	if ext == nil {
		return nil, fmt.Errorf("not defined...")
	}
	return &Column{
		Table: table,
		Def:   ext,
		Field: field,
	}, nil
}

func (c *Column) GetFieldname() string {
	if c.Def.Fieldname != nil {
		return c.Def.GetFieldname()
	}
	return c.GetName()
}

func (c *Column) GetName() string {
	return string(c.Field.GoName)
}

// IsNullable reports whether the column accepts NULL. An explicit `nullable`
// in the column option wins; otherwise it follows the proto field's presence,
// which is the closest proto-native notion of "may be absent": proto2
// optional, proto3 optional and message fields have presence, a proto3 bare
// scalar does not. Primary keys are never nullable.
func (c *Column) IsNullable() bool {
	if c.Def.Nullable != nil {
		return c.Def.GetNullable()
	}
	if c.Def.GetPk() != schemav1.PK_PK_UNSPECIFIED {
		return false
	}
	return c.Field.Desc.HasPresence()
}

func (c *Column) GetSqlType(repo TableRepo, dialect string) (string, error) {
	t, ok := c.Def.Type[dialect]
	if ok {
		return t, nil
	}
	switch c.Field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "BOOLEAN", nil
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		// SQLite requires the literal type name "INTEGER" for a column to be
		// eligible as a ROWID alias, which AUTOINCREMENT (PK_AUTO) requires.
		if dialect == "sqlite3" {
			return "INTEGER", nil
		}
		if c.Field.Desc.Kind() == protoreflect.Int32Kind {
			return "INT(11)", nil
		}
		return "BIGINT", nil
	case protoreflect.StringKind:
		return "VARCHAR(255)", nil
	case protoreflect.FloatKind:
		return "FLOAT", nil
	case protoreflect.MessageKind:
		if c.Def.ForeignKey != nil {
			entity := string(c.Field.Message.Desc.Name())
			var refCols []*Column
			if c.Def.ForeignKey.Entity != nil {
				entity = c.Def.ForeignKey.GetEntity()
			}

			tbl, ok := repo.GetByName(entity)
			if !ok {
				return "", fmt.Errorf("foreign key table '%s' not found", entity)
			}
			for _, ref := range tbl.GetColumns() {
				if len(c.Def.ForeignKey.Fieldnames) != 0 {
					if slices.Contains(c.Def.ForeignKey.Fieldnames, ref.GetName()) {
						refCols = append(refCols, ref)
					}
				} else if ref.Def.GetPk() != schemav1.PK_PK_UNSPECIFIED {
					refCols = append(refCols, ref)
				}
			}
			if len(refCols) == 0 {
				return "", errors.New("no reference column found")
			}
			return refCols[0].GetSqlType(repo, dialect)
		}

		return "JSON", nil
	default:
		return "", errors.New("unknown datatype")
	}
}
