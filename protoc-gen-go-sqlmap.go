package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/generators/plain"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

/* use gogo: */

// func main() {
// 	req := command.Read()
// 	resp := command.GeneratePlugin(req, gogo.New(), ".sqlmap.go")
// 	command.Write(resp)
// }

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		panic(err)
	}

	opts := protogen.Options{}

	builder, err := plain.New(opts, &request)
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
