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

func GenerateImport(name string, importPath string, g *protogen.GeneratedFile) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

func WriteQueries(g Printer, m *types.Table) {
	tplContent, err := f.ReadFile(fmt.Sprintf("sqldialects/%s/select.tmpl", m.Engine))
	if err != nil {
		panic(err)
	}
	tpl, err := template.New("Selects").Funcs(GetTemplateFuns(g)).Parse(string(tplContent))
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteInsertes(g Printer, m *types.Table) {
	tplContent, err := f.ReadFile(fmt.Sprintf("sqldialects/%s/insert.tmpl", m.Engine))
	if err != nil {
		panic(err)
	}
	tpl, err := template.New("Inserts").Funcs(GetTemplateFuns(g)).Parse(string(tplContent))
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteNestedSelectors(g Printer, m *types.Table) {
	tplContent, err := f.ReadFile(fmt.Sprintf("sqldialects/%s/nested.tmpl", m.Engine))
	if err != nil {
		panic(err)
	}
	tpl, err := template.New("Nested").Funcs(GetTemplateFuns(g)).Parse(string(tplContent))
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteUpdates(g Printer, m *types.Table) {
	tplContent, err := f.ReadFile(fmt.Sprintf("sqldialects/%s/update.tmpl", m.Engine))
	if err != nil {
		panic(err)
	}
	tpl, err := template.New("Update").Funcs(GetTemplateFuns(g)).Parse(string(tplContent))
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(g, m)
	if err != nil {
		panic(err)
	}
}

func WriteDeletes(g Printer, m *types.Table) {
	tplContent, err := f.ReadFile(fmt.Sprintf("sqldialects/%s/delete.tmpl", m.Engine))
	if err != nil {
		panic(err)
	}
	tpl, err := template.New("Delete").Funcs(GetTemplateFuns(g)).Parse(string(tplContent))
	if err != nil {
		panic(err)
	}
	err = tpl.Execute(g, m)
	if err != nil {
		panic(err)
	}
}
