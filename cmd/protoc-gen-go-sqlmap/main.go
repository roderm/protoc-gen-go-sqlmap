package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func parseConfig(req *pluginpb.CodeGeneratorRequest) sqlmap.Config {
	cfg := sqlmap.Config{
		Dialect: "postgres", // default
	}

	params := req.GetParameter()
	if params == "" {
		return cfg
	}

	for _, p := range strings.Split(params, ",") {
		kv := strings.SplitN(p, "=", 2)

		switch kv[0] {
		case "dialect":
			if len(kv) == 2 {
				cfg.Dialect = kv[1]
			}
		}
	}

	return cfg
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		panic(err)
	}

	opts := protogen.Options{}

	cfg := parseConfig(&request)
	builder, err := sqlmap.New(opts, &request, cfg)
	if err != nil {
		panic(err)
	}

	response, err := builder.Generate()
	if err != nil {
		panic(err)
	}

	out, err := proto.Marshal(response)
	if err != nil {
		panic(err)
	}

	fmt.Fprint(os.Stdout, string(out))
}
