package plain

import (
	"sort"
	"strings"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/writer"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type SqlGenerator struct {
	plugin         *protogen.Plugin
	messages       map[string]struct{}
	currentPackage string
	tables         *types.TableMessages
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

func (p *SqlGenerator) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	p.loadTables()
	p.loadFields()
	p.loadDependencies()

	writer.TableMessageStore = p.tables
	for _, protoFile := range p.plugin.Files {
		store := p.StoreName(protoFile.GeneratedFilenamePrefix)
		tables := p.tables.GetTablesByStoreName(store)

		if len(tables) == 0 {
			continue
		}
		fileName := protoFile.GeneratedFilenamePrefix + ".sqlmap.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")

		p.currentPackage = protoFile.GoImportPath.String()

		g.P("package ", protoFile.GoPackageName)
		writer.GenerateImport("fmt", "fmt", g)
		writer.GenerateImport("pg", "github.com/roderm/protoc-gen-go-sqlmap/lib/go/pg", g)
		writer.GenerateImport("sql", "database/sql", g)
		writer.GenerateImport("driver", "database/sql/driver", g)
		writer.GenerateImport("context", "context", g)
		writer.GenerateImport("json", "encoding/json", g)

		g.P(`
			var _ = fmt.Sprintf
			var _ = context.TODO
			var _ = pg.NONE
			var _ = sql.Open
			var _ = driver.IsValue
			var _ = json.Valid
		`)
		g.P(`
			type ` + store + ` struct {
				conn *sql.DB
			}
			func New` + store + `(conn *sql.DB) *` + store + ` {
				return &` + store + `{conn}
			}
		`)

		// sort to keep order
		sort.Slice(tables, func(i, j int) bool {
			return tables[i].MsgName < tables[j].MsgName
		})
		for _, table := range tables {
			writer.WriteEncodings(g, table)
			writer.WriteNestedSelectors(g, table)
			writer.WriteQueries(g, table)
			writer.WriteInsertes(g, table)
			writer.WriteUpdates(g, table)
			writer.WriteDeletes(g, table)
			for n, p := range table.Imports {
				writer.GenerateImport(n, p, g)
			}
		}

	}
	return p.plugin.Response(), nil
}
