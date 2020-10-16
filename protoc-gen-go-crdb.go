package main

import (
	"bragi.eyevip.io/rot/protoc-gen-go-sqlmap/sqlgen"
	// _ "bragi.eyevip.io/rot/protoc-gen-go-sqlmap/sqlgen"
	"github.com/gogo/protobuf/vanity/command"
)

func main() {
	req := command.Read()
	resp := command.GeneratePlugin(req, sqlgen.New(), ".crdb-sql.go")
	command.Write(resp)
}
