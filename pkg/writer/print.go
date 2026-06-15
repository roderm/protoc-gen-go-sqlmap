package writer

import (
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"google.golang.org/protobuf/compiler/protogen"
)

type Printer interface {
	P(str ...any)
	Write(p []byte) (n int, err error)
	QualifiedGoIdent(protogen.GoIdent) string
}

type Writer interface {
	Tables(...*types.Table) error
}
