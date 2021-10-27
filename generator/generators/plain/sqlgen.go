package plain

import (
	// pb "github.com/gogo/protobuf/protoc-gen-gogo/descriptor"

	"strings"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/writer"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen/v1"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
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
	plugin         *protogen.Plugin
	messages       map[string]struct{}
	currentFile    string
	currentPackage string
	tables         *types.TableMessages
	// *generator.Generator
	// generator.PluginImports
	// file       *generator.FileDescriptor
	// localName  string
	// atleastOne bool
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*SqlGenerator, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	gen := &SqlGenerator{
		plugin:   plugin,
		messages: make(map[string]struct{}),
		tables: &types.TableMessages{
			MessageTables: make(map[string]*types.Table),
		},
	}
	return gen, nil
}
func (p *SqlGenerator) Name() string {
	return "sqlmap"
}

// Init is called once after data structures are built but before
// code generation begins.
// func (p *SqlGenerator) Init(g *generator.Generator) {
// 	p.Generator = g
// }

func (p *SqlGenerator) SQLDialect() string {
	return "postgres"
}
func (p *SqlGenerator) StoreName(filename string) string {
	path := strings.Split(filename, "/")
	file := strings.Split(path[len(path)-1], ".")
	snake := strings.Split(file[0], "_")
	for i, word := range snake {
		snake[i] = strings.Title(word)
	}
	snake = append(snake, "Store")
	return strings.Join(snake, "")
}

func hasOperation(tbl *sqlgen.Table, op sqlgen.OPERATION) bool {
	for _, o := range tbl.Crud {
		if o == op {
			return true
		}
	}
	return false
}

var genFileMap = make(map[string]*protogen.GeneratedFile)

func generateImport(name string, importPath string, g *protogen.GeneratedFile) string {
	return g.QualifiedGoIdent(protogen.GoIdent{
		GoName:       name,
		GoImportPath: protogen.GoImportPath(importPath),
	})
}

func (p *SqlGenerator) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	p.loadTables()
	p.loadFields()
	p.loadDependencies()

	for _, protoFile := range p.plugin.Files {
		fileName := protoFile.GeneratedFilenamePrefix + ".sqlmap.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")
		store := p.StoreName(protoFile.GeneratedFilenamePrefix)

		p.currentPackage = protoFile.GoImportPath.String()

		g.P("package ", protoFile.GoPackageName)
		generateImport("fmt", "fmt", g)
		generateImport("pg", "github.com/roderm/protoc-gen-go-sqlmap/lib/go/pg", g)
		generateImport("sql", "database/sql", g)
		generateImport("driver", "database/sql/driver", g)
		generateImport("context", "context", g)
		generateImport("json", "encoding/json", g)
		generateImport("binary", "encoding/binary", g)
		g.P(`
			var _ = fmt.Sprintf
			var _ = context.TODO
			var _ = pg.NONE
			var _ = sql.Open
			var _ = driver.IsValue
			var _ = json.Valid
			var _ = binary.LittleEndian
		`)
		g.P(`
			type ` + store + ` struct {
				conn *sql.DB
			}
			func New` + store + `(conn *sql.DB) *` + store + ` {
				return &` + store + `{conn}
			}
		`)
		writer.TableMessageStore = p.tables
		for _, table := range p.tables.GetTablesByStoreName(store) {
			writer.WriteEncodings(g, table)
			writer.WriteNestedSelectors(g, table)
			writer.WriteQueries(g, table)
			writer.WriteInsertes(g, table)
			writer.WriteUpdates(g, table)
			writer.WriteDeletes(g, table)
		}

	}
	return p.plugin.Response(), nil
}
