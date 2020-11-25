package main

import (
	"github.com/gogo/protobuf/vanity/command"
	"github.com/roderm/protoc-gen-go-sqlmap/generator"
)

func main() {
	req := command.Read()
	resp := command.GeneratePlugin(req, generator.New(), ".sqlmap.go")
	command.Write(resp)
}
