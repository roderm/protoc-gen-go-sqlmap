# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`protoc-gen-go-sqlmap` is a `protoc`/`buf` code-generator plugin. Given `.proto` messages annotated with `sqlmap` extensions (table/column metadata, primary keys, foreign keys), it emits a `<file>.sqlmap.go` per proto file containing `schema.v1.SchemaTable`/`SchemaColumn` values plus row-scanning helpers. `pkg/migration` then converts those values into `ariga/atlas` schema types to create tables or write migration files.

The goal is a set of independently switchable plugins over one shared `TableRepo`: (1) atlas migrations — **done**; (2) queries with eager loading of children + FieldMask support — **done**; (3) SQL filters built from the filtrify AST (`github.com/roderm/filtrify`, `proto/filtrify/ast/v1/ast.proto`). Note `docs/plans/sqlmap-generator/PLAN-SQLMAP-GENERATOR.md` is stale on (3): its Phase 5 sketches ad-hoc per-column builder funcs rather than translating an external AST.

Plugins are toggled with the `plugins=` protoc parameter, a **`+`-separated** list (protoc already uses `,` between parameters): `--go-sqlmap_out=plugins=schema+scanner:.`. Omitting it enables everything. `plugins` entries live in one slice in `generator.go`; `query` declares `Requires: ["scanner"]` because its output calls the scanner's `Result` type, and enabling it alone is a generation-time error rather than uncompilable Go.

**Dependency rule: only `ariga.io/atlas` and `google.golang.org/protobuf` are allowed as direct dependencies.** Ask before adding any other, test-only ones included — prefer plain stdlib loops over utility libraries. Target Go 1.27.

## Commands

```bash
task install         # go mod tidy && go install ./cmd/protoc-gen-go-sqlmap/
make install          # buf generate && go install (regenerates pkg/generated/** from proto/)
buf generate          # regenerate pkg/generated/**.pb.go from proto/
go build ./...
go test ./...                                            # golden + unit tests
go test ./pkg/generator/sqlmap/ -update                  # accept new golden output
SQLMAP_E2E=1 go test ./pkg/generator/sqlmap/ -run TestE2E # dockerized-Postgres end-to-end
```

CI (`.github/workflows/test.yaml`) installs protoc + `protoc-gen-go` + buf, runs `make install`, then `go test -run=.`.

### Tests

- `pkg/generator/sqlmap/golden_test.go` — runs `testdata/*.proto` through protoc + the generator in-process and diffs against `testdata/*.sqlmap.go.golden`. `-update` rewrites them.
- `pkg/migration/migration_test.go` — unit tests over `toSchema` (FK pointer identity, auto-increment per dialect, nullability, SET-NULL validation).
- `pkg/query/query_test.go` — mask parsing (including that a broad path beats a narrow one in either order), placeholder renumbering, join-key normalization.
- `pkg/generator/sqlmap/plugins_test.go` — `plugins=` parsing, the scanner dependency, and that a disabled plugin's output really disappears.
- `pkg/generator/sqlmap/e2e_test.go` — opt-in (`SQLMAP_E2E=1`, needs docker). `runE2E` starts a throwaway PostgreSQL container, generates a testdata proto through the real pipeline into a temp Go module, and runs a driver against the live database. Two cases: `relation.proto` (nullability, `ON DELETE SET NULL`) and `eager.proto` (FieldMask column selection, eager loading both directions, two levels deep). They need separate containers because both protos declare `tbl_author`/`tbl_book`. The SQL driver (`lib/pq`) lives in the temp module's `go.mod`, never in this repo's — that's how the e2e keeps the two-dependency rule.

## Architecture

### Generation pipeline

```
.proto file (sqlmap extensions)
    │
    ▼
cmd/protoc-gen-go-sqlmap/main.go       reads CodeGeneratorRequest from stdin, parses `dialect=` param (default "postgres")
    │
    ▼
pkg/generator/sqlmap/generator.go       SqlGenerator.Generate():
    │                                     1. builds a TableRepo by scanning every message for the (sqlgen.table) option
    │                                     2. for each proto file, runs every registered writer against that repo
    ▼
pkg/writer/schema/writer.go              emits schemav1.SchemaTable + SchemaColumn vars, FK wiring via init()
pkg/writer/scanner/writer.go             emits `<Msg>Result` struct with Scan(), GetColValue(), Get<Msg>Columns()/PKColumns()
pkg/writer/query/writer.go               emits <Msg>QueryColumns(), Load<Msg>Rows(), Load<Msg>()
    │
    ▼
<file>.sqlmap.go   (written via protogen.GeneratedFile, package = proto file's go_package)
    │
    ▼
pkg/migration (runtime, not generated)   toSchema() → atlas schema.Table; Create/Diff/ApplyPending
pkg/query     (runtime, not generated)   Mask, Conn/Select, In/Key/Keys placeholder + join-key helpers
```

Writers are plain functions registered in a slice in `generator.go` (`writers: []func(writer.Printer, types.TableRepo) writer.Writer{schema.New, scanner.New}`). Adding a new generated artifact means adding a new writer package + appending it to that slice — no changes needed to the core loop.

Both writers use Go `html/template` (not `text/template`) against small local `Table`/`Column` view-model structs — they translate the richer `types.Table`/`types.Column` domain model into template-friendly shapes right before executing.

### Domain model (`pkg/generator/sqlmap/types`)

- `TableRepo` (`[]*Table`) — all tables across all processed proto files; `GetByName` looks up by message name, used to resolve cross-message foreign keys (including across proto files/packages).
- `Table` wraps a `*protogen.Message` + its `sqlmapv1.Table` extension; `NewTableFromDescriptor` returns an error (not a panic) for messages that don't carry the `(sqlgen.table)` option, and the generator silently skips those — this is the mechanism that decides which messages become tables.
- `Column` wraps a `*protogen.Field` + its `sqlmapv1.Column` extension the same way; fields without `(sqlgen.col)` are skipped.
- `Column.GetSqlType(repo, dialect)` resolves the SQL type: explicit `Def.Type[dialect]` wins; otherwise it falls back by proto kind (bool→BOOLEAN, int32→INT(11), int64→BIGINT, string→VARCHAR(255), float→FLOAT; int32/int64 →INTEGER on sqlite3, which AUTOINCREMENT requires). A `MessageKind` field with a `foreign_key` follows the reference into the target table (via `repo.GetByName`) and recurses into that table's PK (or the FK's explicit `fieldnames`) to inherit its type; a `MessageKind` field *without* a foreign key becomes an embedded `JSON` column. Resolution failures are returned as errors, never baked into the generated file.
- `Column.IsNullable()` — an explicit `nullable` in the column option wins; otherwise PKs are NOT NULL and every other column follows `Field.Desc.HasPresence()` (proto2 `optional`, proto3 `optional`, and message fields have presence; a proto3 bare scalar does not). `pkg/migration` rejects `ON DELETE SET NULL` on a NOT NULL column rather than emitting DDL the database will refuse.
- The schema writer emits a `SchemaType` entry for **every** dialect in its `dialects` var (mysql/postgres/sqlite3), so one generated file serves any of them; the dialect is chosen at runtime by `migration.New`, not at generation time.

### Relations and eager loading (`types/relation.go`, `pkg/writer/query`)

A message-kind field with a `foreign_key` is a **relation**; the field's cardinality picks the direction, so one option syntax covers both:

- **singular** (`optional Publisher publisher = 4`) — belongs-to. It is *both* a column (holding the key) and a relation. The scanner keeps the raw key in `fk_<column>_id` because the Go field is the message the relation will be filled with.
- **repeated** (`repeated Book books = 3`) — has-many. The key lives on the target rows, so it is a relation **only**: `NewTableFromDescriptor` keeps it out of `GetColumns()`, which is what stops it becoming a bogus column and a bogus schema FK.

In both cases `ForeignKey.Fieldnames` names columns on the *target* table, resolved by `types.ResolveRefColumns` (SQL name first, then Go name — both spellings exist in the wild).

Related rows are fetched with one batched `IN (...)` query per relation and stitched in Go, not joined — a join would multiply the parent row out once per child. Because a child load calls the child's own `Load<T>Rows`, nesting is recursive and nested mask paths (`books.publisher.name`) work for free. `extra ...string` forces the join column into the child's SELECT even when the mask omits it.

The FieldMask drives column selection *and* which relations load, so an unmasked relation costs no query. Primary keys are always selected regardless of the mask, since they are what stitching keys on. `query.Key` normalizes join keys because one side comes from a driver (`any`) and the other from a typed getter, and drivers disagree on integer width and `[]byte` vs `string`.

### Proto extensions (`proto/sqlmap/v1/sqlmap.proto`)

Note this is `syntax = "proto2"` — extension fields are `optional`/`required`, and the pattern for "is this set" is `Def.Field != nil` rather than zero-value checks.

- `extend google.protobuf.MessageOptions { optional Table table = 800100; }` — put on a message to make it a table. `Table{name, foreign_keys}`.
- `extend google.protobuf.FieldOptions { optional Column col = 800120; }` — put on a field to make it a column. `Column{fieldname, pk (PK_AUTO|PK_MAN), type (map[dialect]sql_type), foreign_key, nullable}`.
- `ForeignKeyDefinition{fieldnames, to: ForeignKey{entity, fieldnames}, on_delete}` — table-level FK declarations combine with column-level `foreign_key` (via `Table.GetForeignKeys()`, which merges both sources).
- **`README.md`'s usage example is stale** — it shows a `crud: [C,R,D]` field on `Table` that no longer exists in the .proto. Trust the `.proto` file and `pkg/generated/sqlmap/v1/sqlmap.pb.go` over the README when they disagree.

After editing `sqlmap.proto`, regenerate the extension Go types with `buf generate` (per `buf.gen.yaml`) before building — `pkg/generated/sqlmap/v1/sqlmap.pb.go` is generated, not hand-written.

### Writer/Printer abstraction (`pkg/writer/print.go`)

`Printer` is a minimal interface (`P`, `Write`, `QualifiedGoIdent`, `Import`) satisfied by `*protogen.GeneratedFile`, so writer packages depend on it instead of `protogen` directly — keeps `pkg/writer/schema` and `pkg/writer/scanner` testable/mockable without a real `protogen.Plugin`. `Writer` is the per-writer contract: `Write(*protogen.File) error`.

### Cross-file/package foreign keys

When a FK's referenced table lives in a different proto file (`refTable.File.GoImportPath != table.File.GoImportPath`), `schema/writer.go` calls `s.o.QualifiedGoIdent(...)` to get an import-qualified reference instead of a bare identifier — this is the mechanism that makes FKs work across proto packages.
