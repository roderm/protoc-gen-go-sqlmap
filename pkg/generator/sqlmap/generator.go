package sqlmap

import (
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer/scanner"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer/schema"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type SqlGenerator struct {
	plugin  *protogen.Plugin
	writers []func(writer.Printer, types.TableRepo) writer.Writer
	repo    types.TableRepo
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*SqlGenerator, error) {
	plugin, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	gen := &SqlGenerator{
		plugin: plugin,
		writers: []func(writer.Printer, types.TableRepo) writer.Writer{
			schema.New,
			scanner.New,
		},
		repo: make(types.TableRepo, 0),
	}
	return gen, nil
}

func (p *SqlGenerator) Generate() (*pluginpb.CodeGeneratorResponse, error) {
	for _, protoFile := range p.plugin.Files {
		for _, msg := range protoFile.Messages {
			table, err := types.NewTableFromDescriptor(protoFile, msg)
			if err == nil {
				p.repo = append(p.repo, table)
			}
		}
	}
	for _, protoFile := range p.plugin.Files {
		fileName := protoFile.GeneratedFilenamePrefix + ".sqlmap.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")

		g.P("package ", protoFile.GoPackageName)
		for _, writer := range p.writers {
			writer(g, p.repo).Write(protoFile)
		}
	}

	return p.plugin.Response(), nil
}
