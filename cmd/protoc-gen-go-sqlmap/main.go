package main

import (
	"io"
	"os"

	"github.com/roderm/protoc-gen-go-sqlmap/pkg/generator/sqlmap"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

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

	response, err := generate(&request)
	if err != nil {
		// protoc expects a failed plugin to hand back a response carrying the
		// message, so the compiler can report it. Panicking instead surfaces a
		// Go stack trace and a bare "exit status 2" to the user.
		response = &pluginpb.CodeGeneratorResponse{Error: proto.String(err.Error())}
	}

	out, err := proto.Marshal(response)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stdout.Write(out); err != nil {
		panic(err)
	}
}

func generate(request *pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	builder, err := sqlmap.New(protogen.Options{}, request)
	if err != nil {
		return nil, err
	}
	return builder.Generate()
}
