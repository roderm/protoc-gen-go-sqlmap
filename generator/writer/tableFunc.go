package writer

import (
	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

func GetPrimaryKeys(t types.Table) []*types.Field {
	keys := []*types.Field{}
	for _, f := range t.Cols {
		if f.PK != sqlgen.PK_PK_UNSPECIFIED {
			keys = append(keys, f)
		}
	}
	return keys
}

func GetSimpleColumns(t types.Table) []*types.Field {
	c := []*types.Field{}
	for _, f := range t.Cols {
		if f.IsMessage {
			continue
		}
		if f.IsRepeated {
			continue
		}
		c = append(c, f)
	}
	return c
}

func GetMessageColumns(t types.Table) []*types.Field {
	c := []*types.Field{}
	for _, f := range t.Cols {
		if f.IsMessage {
			c = append(c, f)
		}
	}
	return c
}
