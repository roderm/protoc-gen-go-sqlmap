package types

import (
	"errors"
	"fmt"

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

// GetProtoName returns the proto field name, which is what a FieldMask path
// refers to.
func (c *Column) GetProtoName() string {
	return string(c.Field.Desc.Name())
}

// IsPrimaryKey reports whether the column is part of the primary key.
func (c *Column) IsPrimaryKey() bool {
	return c.Def.GetPk() != schemav1.PK_PK_UNSPECIFIED
}

// GetForeignKeyEntity returns the message name the column's foreign key points
// at, defaulting to the field's own message type when `entity` is not spelled
// out -- which is the common case, since a message-kind field already names
// what it references.
func (c *Column) GetForeignKeyEntity() string {
	if c.Def.ForeignKey.GetEntity() != "" {
		return c.Def.ForeignKey.GetEntity()
	}
	if c.Field.Message != nil {
		return string(c.Field.Message.Desc.Name())
	}
	return ""
}

// IsMessage reports whether the column's value is a reference stored as a
// message field, which the scanner keeps as a raw fk_<column>_id value rather
// than scanning into the message itself.
func (c *Column) IsMessage() bool {
	return c.Field.Desc.Kind() == protoreflect.MessageKind && c.Def.ForeignKey != nil
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
		switch dialect {
		case "sqlite3":
			// SQLite requires the literal type name "INTEGER" for a column to
			// be eligible as a ROWID alias, which AUTOINCREMENT (PK_AUTO)
			// requires.
			return "INTEGER", nil
		case "mysql":
			if c.Field.Desc.Kind() == protoreflect.Int32Kind {
				return "INT(11)", nil
			}
			return "BIGINT", nil
		default:
			// PostgreSQL has no display-width syntax: INT(11) is a syntax
			// error there, not a wider integer.
			if c.Field.Desc.Kind() == protoreflect.Int32Kind {
				return "INTEGER", nil
			}
			return "BIGINT", nil
		}
	case protoreflect.StringKind:
		return "VARCHAR(255)", nil
	case protoreflect.FloatKind:
		return "FLOAT", nil
	case protoreflect.MessageKind:
		if c.Def.ForeignKey != nil {
			entity := c.GetForeignKeyEntity()
			tbl, ok := repo.GetByName(entity)
			if !ok {
				return "", fmt.Errorf("foreign key table '%s' not found", entity)
			}
			refCols, err := ResolveRefColumns(tbl, c.Def.ForeignKey.GetFieldnames())
			if err != nil {
				return "", err
			}
			// A foreign key column inherits the type of what it points at.
			return refCols[0].GetSqlType(repo, dialect)
		}

		return "JSON", nil
	default:
		return "", errors.New("unknown datatype")
	}
}
