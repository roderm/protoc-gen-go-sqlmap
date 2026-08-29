# Project Overview

This project contains protoc plugins for generating SQL Mapping. The goal is to define the SQL-schema in protobuf and allow to create ariga/atlas migrations. Further simple bindings for query-builders should be easy: Get rows and types for a table (message) and Scanning to the message.

## Tech Stack
- Go 1.27
- protoc / buf build
- ariga.io/atlas (schema diffing, DDL planning, migrations)
- google.golang.org/protobuf

Only those two Go modules are permitted as direct dependencies — ask before adding another, test-only ones included.

## Architecture

### High-Level Flow
```
.proto file (with sqlmap extensions)
    |
    v
protoc-gen-go-sqlmap (this binary)
   +- Parse proto descriptors & extensions
   +- Build TableRepo from messages with (sqlgen.table) option
   +- Extract columns with (sqlgen.col) field extensions
   +-- Generate Go code via template-driven writers
        |
        +- SchemaWriter   -> SQL schema/table definitions (schemav1.SchemaTable, consumed by pkg/migration -> atlas)
        +-- ScannerWriter  -> Result struct, Scan(), GetColValue(), column helpers
```

### Directory Layout
| Path                                    | Purpose                                                                                              |
|-----------------------------------------|------------------------------------------------------------------------------------------------------|
| `cmd/protoc-gen-go-sqlmap/main.go`      | Entry-point; parses `dialect=` parameter & invokes the generator                                     |
| `pkg/generator/sqlmap/generator.go`     | Core SqlGenerator: collects Tables, dispatches to writers                                            |
| `pkg/generator/sqlmap/types/table.go`   | TableRepo + Table type holding proto-derived table metadata                                          |
| `pkg/generator/sqlmap/types/column.go`  | Column type with SQL-type resolution per dialect (Postgres/MySQL) & FK lookups                      |
| `pkg/writer/schema/writer.go`           | Template -> generates schema.Table + column definitions                                              |
| `pkg/writer/scanner/writer.go`          | Template -> generates Result struct with Scan() & helpers                                            |
| `pkg/writer/print.go`                   | Printer / Writer interface (small abstraction over protogen.File)                                    |
| `./proto/sqlmap/v1/sqlmap.proto`        | Extension definitions (table, col, ForeignKey, PK enum, OnDelete)                                    |
| `buf.yaml`                              | Buf module config, lint rules, breaking checks                                                       |
| `buf.gen.yaml`                          | Buf generation: go-pb + (optional SQL dialect plugins)                                               |
| `Taskfile.yml` / `Makefile`             | Dev convenience: generate & install                                                                    |

### Key Types & Interfaces

```go
// pkg/generator/sqlmap
type Config struct { Dialect string }           // "postgres" default
type SqlGenerator struct { /* ... */ }
func (g *SqlGenerator) Generate() (*pluginpb.CodeGeneratorResponse, error)

// pkg/generator/sqlmap/types
type TableRepo []*Table                          // slice of resolved tables
type Table struct { Def *sqlmapv1.Table; Msg *protogen.Message; columns []*Column }
func NewTableFromDescriptor(f *protogen.File, msg *protogen.Message) (*Table, error)

// pkg/generator/sqlmap/types/column.go
type Column struct { Def *sqlmapv1.Column; Field *protogen.Field }
func (c *Column) GetSqlType(repo TableRepo, dialect string) (string, error)

// pkg/writer
type Printer  interface { P(); Write(); ... }
type Writer  interface { Write(*protogen.File) error } // implemented by SchemaWriter & ScannerWriter
```

### SQL Type Resolution (pkg/generator/sqlmap/types/column.go)

When a Column's SQL type is not specified in the proto extension, `GetSqlType()` falls back to Go/Proto kind:

| Proto Kind      | Default PostgreSQL | Default MySQL (ent dialect) | Notes                              |
|-----------------|--------------------|-----------------------------|------------------------------------|
| Bool            | BOOLEAN            | BOOLEAN                     |                                    |
| Int32           | INT(11)            |                             |                                    |
| Int64           | BIGINT             |                             |                                    |
| String          | VARCHAR(255)       |                             | Default fallback for strings       |
| Float           | FLOAT              |                             |                                    |
| Message (FK)    | resolves FK target | resolves FK target          | FollowsForeignKeyDef to find type  |
| Message (no FK) | JSON               | JSON                        | Embedded JSONB                     |

### Configuration via protobuf Extensions

Extension definitions are in `./proto/sqlmap/v1/sqlmap.proto`. Key enum values:
- `PK_UNSPECIFIED` -> not a PK
- `PK_AUTO`       -> auto-incrementing primary key  
- `PK_MAN`        -> manual (user-supplied) primary key
- `ON_DELETE_CASCADE`, `ON_DELETE_SET_NULL`

```protobuf
// On a message: define the table
option (sqlgen.table) = { name: "tbl_name" };

// On a field: define the column  
field_name type TypeName = 1 [(sqlgen.col) = {
  fieldname: "db_column_name"   // overrides Go-name for SQL; defaults to Go Name
  pk: PK_AUTO | PK_MAN           // primary key, AUTO generates with sql.AutoIncrement()
  type: { postgres: "VARCHAR", mysql: "varchar(255)" }
  foreign_key { entity: "OtherMessage", fieldnames: ["id"] on_delete: ON_DELETE_SET_NULL }
}];
```

### Generated Code Pattern (.sqlmap.go)

For each `.proto` file that contains `option(sqlgen.table)` messages, the output `.sqlmap.go`:

1. Declares column variables:  
   `EntityNameColumnDef_FieldName = &schema.Column{Name: "...", SchemaType: map[string]string{...}}`
2. Creates a table variable:  
   `var EntityTableName *schema.Table = ...`
3. Generates `init()` to attach foreign keys
4. Generates ScannerWriter output per message:
   - `type EntityResult struct { Entity }` with auto-generated `Scan(cols, r)` method
   - `GetEntityPKColumns() []string`, `GetEntityColumns() []string`
   - `GetEntityColValue(col string) any`

### Generator Flow Summary

1. `main()` reads stdin (`pluginpb.CodeGeneratorRequest`) -> parses `dialect` param (default "postgres")
2. `SqlGenerator.Generate()` iterates all proto files:  
   - For each message with `table` option: builds `Table` via `NewTableFromDescriptor()` -> populates `TableRepo`  
   - For each message with `table` option: calls every registered writer (schema, scanner) with the repo
3. Writers use Go text templates to emit structured code into generated `.sqlmap.go`
4. Returns `pluginpb.CodeGeneratorResponse` via stdout

## Development Commands

| What                                          | Command                     |
|-----------------------------------------------|-----------------------------|
| Install binary                                | `task install` or `make install` |
| Proto + Buf generate                          | `buf generate`              |
| Full regenerate (clones protobuf into vendor/) | `make regenerate`          |
| Run tests                                     | `go test -run=.`            |

## Testing Notes

- `go test ./...` runs golden-file tests (`pkg/generator/sqlmap`) and `toSchema` unit tests (`pkg/migration`)
- `go test ./pkg/generator/sqlmap/ -update` rewrites `testdata/*.sqlmap.go.golden` from current output
- `SQLMAP_E2E=1 go test ./pkg/generator/sqlmap/ -run TestE2E` runs the opt-in end-to-end test against a throwaway PostgreSQL container (needs docker). Its SQL driver lives in the temp module the test assembles, so it never enters this repo's `go.mod`.
- Lint rules in `buf.yaml`: MINIMAL + additional ENUM/FIELD/PACKAGE style checks

## Roadmap / Open Items (from README.md)

**Queries:**
- [ ] Add parameter to not resolve references (problem with json-encoding)

**Insert:**
- [x] Insert single message to table  
- [ ] Resolve foreign-key (and auto-create on insert)

**Delete:** Done [x]

**Update:**
- [ ] Implement update with replace and conflict handling

**Platform / Feature Gaps:**
- [ ] Support multiple PKs
- [ ] Add multiple FKs for single Message
- [ ] Support `oneOf` type in proto3
- [ ] Improve filter to use table/alias column names (not raw SQL names)
- [ ] Load external messages to resolve foreign keys [x Done]
- [x] Decouple from gogo protobuf [Done]
- [ ] Write generator test [x Done]

**Database Dialects:**
- [x] PostgreSQL
- [ ] CockroachDB
- [ ] MySQL / MariaDB  
- [ ] MSSQL
- [ ] SQLite

## Publishing / Releases

Release process via `.github/workflows/release.yaml`:
1. Tag push (`vX.Y.Z`) triggers the workflow
2. Pushes protos to buf.build (`buf-push-action`)
3. Creates GitHub Release with binaries for linux/amd64, linux/386, windows/amd64, windows/386
