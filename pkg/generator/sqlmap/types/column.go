package types

import (
	"errors"

	sqlmapv1 "github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Column struct {
	table *Table
	def   *sqlmapv1.Column
	field *protogen.Field
}

func NewColumn(table *Table, field *protogen.Field) *Column {
	ext := proto.GetExtension(field.Desc.Options(), sqlmapv1.E_Col).(*sqlmapv1.Column)
	if ext == nil {
		return nil
	}
	return &Column{
		table: table,
		def:   ext,
		field: field,
	}
}

func (c *Column) GetName() string {
	if c.def.Fieldname != nil {
		return *c.def.Fieldname
	}
	return string(c.field.Desc.Name())
}

func (c *Column) GetSqlType(dialect string) (string, error) {
	t, ok := c.def.Type[dialect]
	if ok {
		return t, nil
	}
	switch c.field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "BOOLEAN", nil
	case protoreflect.Int32Kind:
		return "INT(11)", nil
	case protoreflect.Int64Kind:
		return "BIGINT", nil
	case protoreflect.StringKind:
		return "VARCHAR(255)", nil
	case protoreflect.FloatKind:
		return "FLOAT", nil
	case protoreflect.MessageKind:
		return "", errors.New("message kind not supported (yet)")
	default:
		return "", errors.New("unknown datatype")
	}
}
