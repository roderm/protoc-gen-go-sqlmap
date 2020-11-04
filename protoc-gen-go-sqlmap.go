package main

import (
	"github.com/gogo/protobuf/vanity/command"
	"github.com/roderm/protoc-gen-go-sqlmap/sqlgen"
)

func main() {
	req := command.Read()
	resp := command.GeneratePlugin(req, sqlgen.New(), ".sqlmap.go")
	command.Write(resp)
}
