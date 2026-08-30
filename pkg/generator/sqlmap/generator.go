package sqlmap

import (
	"fmt"
	"slices"
	"strings"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap/types"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer/query"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer/scanner"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/writer/schema"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// plugin is one generated artifact. Adding a feature means adding an entry
// here; the core loop needs no changes.
//
// Order matters: writers append to the same file in this order, and the query
// writer's output calls the Result type the scanner writer declares.
type plugin struct {
	Name string
	New  func(writer.Printer, types.TableRepo) writer.Writer
	// Requires names plugins whose output this one refers to.
	Requires []string
}

var plugins = []plugin{
	{Name: "schema", New: schema.New},
	{Name: "scanner", New: scanner.New},
	{Name: "query", New: query.New, Requires: []string{"scanner"}},
}

type SqlGenerator struct {
	plugin  *protogen.Plugin
	writers []func(writer.Printer, types.TableRepo) writer.Writer
	repo    types.TableRepo
}

func New(opts protogen.Options, request *pluginpb.CodeGeneratorRequest) (*SqlGenerator, error) {
	enabled, err := parseEnabled(request.GetParameter())
	if err != nil {
		return nil, err
	}
	plug, err := opts.New(request)
	if err != nil {
		return nil, err
	}
	gen := &SqlGenerator{
		plugin: plug,
		repo:   make(types.TableRepo, 0),
	}
	for _, p := range plugins {
		if !enabled[p.Name] {
			continue
		}
		for _, req := range p.Requires {
			if !enabled[req] {
				return nil, fmt.Errorf("sqlmap: plugin %q needs %q, which is not enabled", p.Name, req)
			}
		}
		gen.writers = append(gen.writers, p.New)
	}
	return gen, nil
}

// parseEnabled reads the `plugins=` protoc parameter, a `+`-separated list of
// plugin names (`+` rather than `,`, which protoc already uses to separate
// parameters). Every plugin is enabled when the parameter is absent.
func parseEnabled(param string) (map[string]bool, error) {
	enabled := make(map[string]bool, len(plugins))
	var spec string
	for _, part := range strings.Split(param, ",") {
		if name, value, ok := strings.Cut(part, "="); ok && strings.TrimSpace(name) == "plugins" {
			spec = value
		}
	}
	if spec == "" {
		for _, p := range plugins {
			enabled[p.Name] = true
		}
		return enabled, nil
	}
	known := make([]string, len(plugins))
	for i, p := range plugins {
		known[i] = p.Name
	}
	for _, name := range strings.Split(spec, "+") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if !slices.Contains(known, name) {
			return nil, fmt.Errorf("sqlmap: unknown plugin %q in plugins=, want any of %s", name, strings.Join(known, "+"))
		}
		enabled[name] = true
	}
	return enabled, nil
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
		// Emit nothing for a file that declares no tables. Well-known types
		// are in the request too, and generating a bare `package` stub for
		// each one both litters the output and breaks the build: with
		// paths=source_relative, descriptor.proto and timestamp.proto land in
		// the same directory under different Go package names.
		if len(p.repo.ForFile(protoFile)) == 0 {
			continue
		}
		fileName := protoFile.GeneratedFilenamePrefix + ".sqlmap.go"
		g := p.plugin.NewGeneratedFile(fileName, ".")

		g.P("package ", protoFile.GoPackageName)
		for _, w := range p.writers {
			if err := w(g, p.repo).Write(protoFile); err != nil {
				return nil, fmt.Errorf("%s: %w", protoFile.Desc.Path(), err)
			}
		}
	}

	return p.plugin.Response(), nil
}
