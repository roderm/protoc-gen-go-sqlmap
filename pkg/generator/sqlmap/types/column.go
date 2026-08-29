package types

import (
	"errors"
	"fmt"

	schemav1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1"
	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"github.com/samber/lo"
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
			if len(c.Def.ForeignKey.Fieldnames) != 0 {
				refCols = lo.Filter(tbl.GetColumns(), func(ref *Column, _ int) bool {
					return lo.Contains(c.Def.ForeignKey.Fieldnames, string(ref.GetName()))
				})
			} else {
				refCols = lo.Filter(tbl.GetColumns(), func(c *Column, _ int) bool {
					return c.Def.Pk != nil && c.Def.Pk != schemav1.PK_PK_UNSPECIFIED.Enum()
				})
			}
			col, ok := lo.First(refCols)
			if !ok {
				return "", errors.New("no reference column found")
			}
			return col.GetSqlType(repo, dialect)
		}

		return "JSON", nil
	default:
		return "", errors.New("unknown datatype")
	}
}
