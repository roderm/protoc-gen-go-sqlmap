package writer

import (
	"google.golang.org/protobuf/compiler/protogen"
)

type Printer interface {
	P(str ...any)
	Write(p []byte) (n int, err error)
	QualifiedGoIdent(protogen.GoIdent) string
	Import(protogen.GoImportPath)
}

type Writer interface {
	Write(*protogen.File) error
}
