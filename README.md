# protoc-gen-go-sqlmap:
This is a protobuf-plugin based on [gogo](https://github.com/gogo/protobuf) to generate SQL-Select mappers in go. In future insert and updates should also be supported.

## Usage:
* Clone this repo (churrently you should checkout `dev`)
* use the `make install` to install `protoc-gen-go-sqlmap` into you're $GOPATH
* in `protoc` you can now use `--go-sqlmap_out=.`

To make use of this plugin, add following extensions to you .proto-file:
```
import "github.com/roderm/protoc-gen-go-sqlmap/sqlgen/sqlgen.proto" // Import of the extensions

message Some {
    option (sqlgen.dbtable) = "tbl_some"; // table where this data ist stored
    string Id = 1 [(sqlgen.dbpk) = AUTO, (sqlgeb.dbcol) = "some_id"]; // It's an autogenerated key in the field "some_id"
    repeated Attribute Attributes = 2; [(sqlgeb.dbcol) = "some_id", (sqlgen.dbfk) = "tbl_some_attributes.some_id"]; // Make a one-to-many link to load
}

message Attribute {
    option (sqlgen.dbtable) = "tbl_some_attributes";
    string Id = 1 [(sqlgen.dbpk) = AUTO, (sqlgeb.dbcol) = "attributes_id"];
    string Name = 2 [(sqlgeb.dbcol) = "attributes_name"];
    bytes Value = 3 [(sqlgeb.dbcol) = "attributes_value"];
    Some Parent = 4 [(sqlgeb.dbcol) = "some_id", (sqlgeb.dbfk) = "tbl_some.some_id"]; // Make a one-to-one link to load
}
```
After generation of the .sqlmap.go-file you should be able to load the data as following:
```
store = package.NewStore(db_conn)
rows, err = store.Some(context.TODO(), package.SomeWithAttributes())
```
Additional there can be given extra parameters for filtering (Currently only forks for Postgres/Cockroach):
`store.Some(context.TODO(),SomeFilter(pg.EQ("some_id", "some_id"))`
or for the sub-query:
`store.Some(context.TODO(), package.SomeWithAttributes(AttributeFilter(pg.EQ("name", "FirstArgument")))`

## Roadmap:
- [] Implement loading of one-to-many
- [] Implement loading of one-to-many
- [] Write tests for querying
- [] Add multiple FKs for single Message
- [] Implement insert
- [] Implement update
- [] Support `oneOf` type in proto3
- [] Implement other Database-Syntax
    - [] CockroachDB
    - [] PostgreSQL
    - [] MySQL / MariaDB

## Gotcha's
* Proto Import: `github.com/gogo/protobuf/protobuf/google/protobuf/descriptor.proto`
* actual go-package: `github.com/gogo/protobuf/protoc-gen-gogo/descriptor`
=> fixed with `gofast_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:.`