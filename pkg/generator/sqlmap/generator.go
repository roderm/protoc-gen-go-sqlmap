package sqlmap

import (
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer/schema"
	"github.com/samber/lo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type Config struct {
	Dialect string
}

type SqlGenerator struct {
	plugin  *protogen.Plugin
	dialect string
	writers []func(writer.Printer, string) writer.Writer
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest, config Config) (*SqlGenerator, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	gen := &SqlGenerator{
		plugin:  plugin,
		dialect: config.Dialect,
		writers: []func(writer.Printer, string) writer.Writer{
			schema.New,
		},
	}
	return gen, nil
}

func (p *SqlGenerator) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	for _, protoFile := range p.plugin.Files {
		fileName := protoFile.GeneratedFilenamePrefix + ".sqlmap.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")

		g.P("package ", protoFile.GoPackageName)
		g.Import("entgo.io/ent/dialect/sql/schema")
		tables := lo.FilterMap(protoFile.Messages, func(msg *protogen.Message, _ int) (*types.Table, bool) {
			table, err := types.NewTableFromDescriptor(protoFile, msg)
			return table, err == nil
		})
		for _, writer := range p.writers {
			writer(g, p.dialect).Tables(tables...)
		}
	}

	return p.plugin.Response(), nil
}
