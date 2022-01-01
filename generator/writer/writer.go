package writer

import (
	"embed"
	"fmt"
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"google.golang.org/protobuf/compiler/protogen"
)

//go:embed sqldialects/*

var f embed.FS

func getTemplate(engine, filename string) *template.Template {
	tplContent, err := f.ReadFile(fmt.Sprintf("sqldialects/%s/%s.go.tpl", engine, filename))
	if err != nil {
		panic(err)
	}
	tpl, err := template.New(filename).Funcs(GetTemplateFuns()).Parse(string(tplContent))
	if err != nil {
		panic(err)
	}
	return tpl
}
func GenerateImport(name string, importPath string, g *protogen.GeneratedFile) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

func WriteQueries(g Printer, m *types.Table) {
	err := getTemplate(m.Engine, "select").Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteInsertes(g Printer, m *types.Table) {
	err := getTemplate(m.Engine, "insert").Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteNestedSelectors(g Printer, m *types.Table) {
	err := getTemplate(m.Engine, "nested").Execute(g, m)
	if err != nil {
		panic(err)
	}
	err = getTemplate(m.Engine, "subloader").Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteUpdates(g Printer, m *types.Table) {
	err := getTemplate(m.Engine, "update").Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteDeletes(g Printer, m *types.Table) {
	err := getTemplate(m.Engine, "delete").Execute(g, m)
	if err != nil {
		panic(err)
	}
}
