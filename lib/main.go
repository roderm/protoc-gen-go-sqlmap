package main

import (
	"fmt"

	"github.com/roderm/protoc-gen-go-sqlmap/lib/pg"
)

func main() {
	eq := pg.AND(pg.NONE(), pg.EQ("test", "bla"))
	fmt.Println(pg.GetWhereClause(eq))
}
