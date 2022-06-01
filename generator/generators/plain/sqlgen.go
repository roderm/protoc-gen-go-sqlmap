package plain

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
	"strings"

	"github.com/fatih/structtag"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
	"github.com/roderm/protoc-gen-go-sqlmap/generator/writer"
	sqlgen "github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
	tagger "github.com/srikrsna/protoc-gen-gotag/module"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

type SqlGenerator struct {
	plugin         *protogen.Plugin
	messages       map[string]struct{}
	currentPackage string
	tables         *types.TableMessages
	params         map[string]string
	generated      map[string]bool
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*SqlGenerator, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	for _, p := range strings.Split(request.GetParameter(), ",") {
		kv := strings.Split(p, "=")
		params[kv[0]] = strings.Join(kv[1:], "=")
	}
	gen := &SqlGenerator{
		plugin:   plugin,
		messages: make(map[string]struct{}),
		tables: &types.TableMessages{
			MessageTables: make(map[string]*types.Table),
		},
		params:    params,
		generated: make(map[string]bool),
	}
	return gen, nil
}

func (p *SqlGenerator) Name() string {
	return "sqlmap"
}

func (p *SqlGenerator) SQLDialect() string {
	if g, ok := p.params["engine"]; ok {
		switch g {
		case "mysql", "mariadb":
			return "mysql"
		}
	}
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
	for _, pf := range p.plugin.Files {
		p.addTags(pf)
		p.createMappers(pf)
	}
	return p.plugin.Response(), nil
}

func (p *SqlGenerator) addTags(pf *protogen.File) {
	fs := token.NewFileSet()
	pbFile := pf.GeneratedFilenamePrefix + ".pb.go"
	if out, ok := p.params["out"]; ok {
		pbFile = fmt.Sprintf("%s/%s", out, pbFile)
	}
	if exists, ok := p.generated[pbFile]; ok && exists {
		return
	} else {
		p.generated[pbFile] = true
	}
	tags := map[string]map[string]*structtag.Tags{}
	for _, f := range pf.Messages {
		tags[string(f.Desc.Name())] = map[string]*structtag.Tags{}
		for _, c := range f.Fields {
			ft := &structtag.Tags{}
			ext, ok := proto.GetExtension(c.Desc.Options(), sqlgen.E_Col).(*sqlgen.Column)
			if !ok || ext == nil {
				continue
			}
			ft.Set(&structtag.Tag{
				Key:     "db",
				Name:    ext.GetName(),
				Options: []string{},
			})

			tags[string(f.Desc.Name())][c.GoName] = ft
		}
	}
	fn, err := parser.ParseFile(fs, pbFile, nil, parser.ParseComments)
	if err != nil {
		// err = fmt.Errorf("%s.pb.go: params[%s => %s] %s", pf.GeneratedFilenamePrefix, pf.GoImportPath, p.plugin.Request.GetParameter(), err)
		// p.plugin.Error(err)
		return
	}
	err = tagger.Retag(fn, tags)
	if err != nil {
		p.plugin.Error(err)
		return
	}
	out := p.plugin.NewGeneratedFile(pf.GeneratedFilenamePrefix+".pb.go", ".")
	printer.Fprint(out, fs, fn)
	if err != nil {
		p.plugin.Error(err)
	}
}

func (p *SqlGenerator) createMappers(pf *protogen.File) {
	store := p.StoreName(pf.GeneratedFilenamePrefix)
	tables := p.tables.GetTablesByStoreName(store)
	if len(tables) == 0 {
		return
	}
	fileName := pf.GeneratedFilenamePrefix + ".sqlmap.go"
	if exists, ok := p.generated[fileName]; ok && exists {
		return
	} else {
		p.generated[fileName] = true
	}
	g := p.plugin.NewGeneratedFile(fileName, ".")

	p.currentPackage = pf.GoImportPath.String()

	g.P("package ", pf.GoPackageName)
	writer.GenerateImport("fmt", "fmt", g)
	writer.GenerateImport("sqlx", "github.com/jmoiron/sqlx", g)
	writer.GenerateImport("driver", "database/sql/driver", g)
	writer.GenerateImport("context", "context", g)
	writer.GenerateImport("json", "encoding/json", g)
	writer.GenerateImport("squirrel", "github.com/Masterminds/squirrel", g)
	writer.GenerateImport("squirrel", "github.com/roderm/gotools/squirrel", g)
	g.P(`
			var _ = fmt.Sprintf
			var _ = context.TODO
			var _ = driver.IsValue
			var _ = json.Valid
			var _ = squirrel.Select
			var _ = sqlx.Connect
			var _ = squirrel1.EqCall{}
		`)
	g.P(`
			type ` + store + ` struct {
				conn *sqlx.DB
			}
			func New` + store + `(conn *sqlx.DB) *` + store + ` {
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
