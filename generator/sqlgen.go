package generator

import (
	// pb "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"fmt"
	"strings"

	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/gogopars"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/writer"
)

type pk string

const (
	pkNone = ""
	pkAuto = "auto"
	pkMan  = "man"
)

// func init() {
// 	generator.RegisterPlugin(New())
// }

type SqlGenerator struct {
	*generator.Generator
	generator.PluginImports
	file       *generator.FileDescriptor
	localName  string
	atleastOne bool
}

func New() generator.Plugin {
	return new(SqlGenerator)
}
func (p *SqlGenerator) Name() string {
	return "sqlmap"
}

// Init is called once after data structures are built but before
// code generation begins.
func (p *SqlGenerator) Init(g *generator.Generator) {
	p.Generator = g
}

func (p *SqlGenerator) StoreName() string {
	path := strings.Split(p.file.GetName(), "/")
	file := strings.Split(path[len(path)-1], ".")
	snake := strings.Split(file[0], "_")
	for i, word := range snake {
		snake[i] = strings.Title(word)
	}
	snake = append(snake, "Store")
	return strings.Join(snake, "")
}

// Generate produces the code generated by the plugin for this file,
// except for the imports, by calling the generator's methods P, In, and Out.
func (p *SqlGenerator) Generate(file *generator.FileDescriptor) {
	p.localName = generator.FileName(file)
	p.PluginImports = generator.NewPluginImports(p.Generator)
	p.file = file

	writer.TableMessageStore = gogopars.NewTableMessages(p.file.Messages())

	p.AddImport(generator.GoImportPath("github.com/roderm/protoc-gen-go-sqlmap/lib/pg"))
	p.AddImport(generator.GoImportPath("database/sql"))
	p.AddImport(generator.GoImportPath("database/sql/driver"))
	p.AddImport(generator.GoImportPath("context"))
	p.AddImport(generator.GoImportPath("encoding/json"))

	fmt.Fprint(p, `
	var _ = context.TODO
	var _ = pg.NONE
	var _ = sql.Open
	var _ = driver.IsValue
	var _ = json.Valid
	`)

	p.StoreName()
	p.P(`
		type ` + p.StoreName() + ` struct {
			conn *sql.DB
		}

		func New` + p.StoreName() + `(conn *sql.DB) *` + p.StoreName() + ` {
			return &` + p.StoreName() + `{conn}
		}
	`)
	for _, tbl := range writer.TableMessageStore.MessageTables {
		writer.WriteConfigStructs(p, tbl)
		writer.WriteQueries(p, tbl)
		writer.WriteUpdates(p, tbl)
		writer.WriteInsertes(p, tbl)
		writer.WriteDeletes(p, tbl)
	}
	if !p.atleastOne {
		return
	}
}
