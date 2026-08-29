# PLAN: protoc-gen-go-sqlmap — Multi-Dialect SQL Generator Roadmap

## Overview

This plan details the work to extend `protoc-gen-go-sqlmap` from its current state (single-dialect PostgreSQL schema + scanner) into a multi-dialect generator with full CRUD support, robust type mapping, and configurable behavior.

## Current State

### What Exists

| Component                    | Status     | Notes                                          |
|------------------------------|------------|------------------------------------------------|
| SchemaWriter                 | Stable     | Generates `schema.Table`, column defs (Postgres/MySQL ent types) |
| ScannerWriter                | Stable     | Generates Result struct, `Scan()`, helper functions |
| TableRepo / Table / Column   | Usable     | Reads proto extensions, resolves SQL types      |
| Config / parseConfig         | Working    | Single `dialect` parameter                      |
| Entry point (main.go)        | Working    | Stdin → protobuf response                        |

### Known Gaps (from README)

- Insert: FK auto-create not implemented
- Update: Not implemented
- Multi-PK / Multi-FK support: Not implemented
- Optional resolution toggle: Not implemented
- `oneOf` in proto3: Not supported
- Naming conventions: No table/column alias filters for raw SQL names
- Dialects: PostgreSQL only (fully), MySQL partially, others zero

---

## Dependency Graph

```
proto/sqlmap/v1/sqlmap.proto  ──→  pkg/generated/sqlmap/v1/*.pb.go   [codegen target — no source changes]
        │                                                    │
        ▼                                                    │
┌──────────────────────┐                                    │
│ cmd/main.go           │──── parses CLI param (dialect) ──┤
└──────────────────────┘                                    │
        │                                                   │
        ▼                                                   │
┌──────────────────────┐    builds TableRepo ──────────────┤
│ generator/sqlmap/     │         │                         │
│   generator.go        │<────────┘                         │
│                       │   for each message with           │
│   Config              │      [sqlgen.table] option        │
│   SqlGenerator        │                                   │
└──────────┬────────────┘                                    │
           │ writes .sqlmap.go per proto file               │
           ▼                                                │
    ┌──────────────────────────┐                            │
    |   writers/                │<────────── dispatch ──────┤
    |   ├── schema/writer.go    │   (registered in New())  │
    |   └── scanner/writer.go  │                           │
    └──────────────────────────┘                            │
           │                                              │
     writes proto file                                     │
```

### Explicit Dependencies Between Components

| Component                         | Depends On                        | Used By              |
|-----------------------------------|------------------------------------|----------------------|
| `main.go`                         | `Config`, `SqlGenerator`          | —                    |
| `generator.go`                    | `types.TableRepo`, all writers     | `main.go`           |
| `types.table.go`                  | `pkg/generated/sqlmap/v1` (pb types) | `generator.go`    |
| `types.column.go`                 | `types.TableRepo`, `sqlmapv1`      | `schema/writer.go`, `scanner/writer.go` |
| `writer/schema/writer.go`         | `types.Column.GetSqlType()`, `ent/dialect/sql/schema` | output → generated file |
| `writer/scanner/writer.go`        | `types.Column.GetName()`, `sqlmapv1.PK_*` | output → generated file |

### Writer Dispatch Pattern (current)

```go
// In generator.go — writers are registered via slice:
gen := &SqlGenerator{
    ...
    writers: []func(writer.Printer, types.TableRepo) writer.Writer{
        schema.New,
        scanner.New,
    },
    ...
}
```

Adding a new feature = adding a new func to this slice. No changes needed in the core loop.

---

## Vertical Slices (Phases)

Each phase delivers **complete, independently testable functionality** for all table/entity types — not horizontal layer-by-layer changes.

---

### Phase 0: Foundation & Test Infrastructure

> **Goal**: Establish reliable test harness before any feature work. Prevent regressions on every slice.

#### Why First?
The current codebase has zero tests despite the roadmap claiming "Write generator test [done]". Every subsequent phase needs regression coverage to safely introduce breaking changes (new proto extensions, type resolution shifts, etc.).

#### Tasks

**T0.1: Proto extension definition audit**
- Review `proto/sqlmap/v1/sqlmap.proto` completeness against all planned features
- Identify where new fields/enums are needed (e.g., `insert_on_conflict`, `oneof_mapping`)
- Output: list of required proto field additions

**T0.2: Generator test harness**
- Create `cmd/protoc-gen-go-sqlmap/internal/testutil` package with helpers to simulate CodeGeneratorRequest, invoke generator, and capture output
- Use `protogen.Testing` helper from the protobuf compiler API
- Output: Testable `RunGenerator(request *pluginpb.CodeGeneratorRequest) (*protogen.Response, error)` function

**T0.3: Golden-file test for Schema + Scanner (baseline)**
- Create test proto file with minimal `table`/`col` extensions
- Run generator → capture output → save as golden file
- Verify output compiles (go vet / go build on generated temp file)
- Output: 1 passing golden test, baseline coverage ~0% → target ~60%

#### Acceptance Criteria
- [ ] `make test` or equivalent runs the generator test harness
- [ ] At least one golden-file test passes and regenerates correctly
- [ ] Test proto files live in `testdata/protos/` with clear extension patterns
- [ ] No regression in existing schema/scanner output for default postgres dialect

#### Checkpoint: Phase 0 Complete When All 3 Tasks Pass ✓ Before proceeding to feature work.

---

### Phase 1: Insert CRUD + Update CRUD (Complete Write Operations)

> **Goal**: Generate `Create[T]`, `Update[T]`, `InsertOrReplace[T]` functions in `.sqlmap.go` with configurable conflict handling for PostgreSQL's INSERT ... ON CONFLICT pattern.

#### Why This Second?
The generator already produces Table metadata (schema writer). Insert/update build directly on top of that — same column list, same table names. Low risk, high value. The README marks Insert as "[x] Insert single message to table" but the code only produces schema + scanner; update is completely absent.

#### Tasks

**T1.1: Proto extension additions for DML operations**
- Add `conflict_target` field to `Column` (postgres upsert columns)
- Add `on_conflict_action` enum (`ON_CONFLICT_NONE`, `ON_CONFLICT_REPLACE`, `ON_CONFLICT_SET_NULL`)
- Add optional `insert_only` boolean flag on Table (skip updates, insert-only mode)

**T1.2: Generate Insert functions**
- New writer: `writer/dml/writer.go` with methods for each entity type
- Generates: `func (*Store) Create(ctx context.Context, t *T) error`
  - Builds column list from `GetColumns()` excluding auto PKs (for insert)
  - Uses INSERT + RETURNING for single-row creation
  - Populates primary key on return
- Generates: `func (*Store) InsertOrReplace(ctx context.Context, t *T) error`
  - PostgreSQL: `INSERT ... ON CONFLICT (...) DO UPDATE SET ...`
  - Columns derived from `conflict_target` in proto or all non-auto PKs

**T1.3: Generate Update functions**
- Generates: `func (*Store) Update(ctx context.Context, t *T) error`
  - Where clause built from all primary key columns (supports multi-PK)
  - SET clause uses all non-PK columns
  - Returns affected row count or error
- Generates: `func (*Store) UpdatePartial(ctx context.Context, t *T, cols []string) error`
  - Partial update for selective column updates

**T1.4: Foreign-key auto-create on Insert**
- Generate companion inserts when entity has FK references to other entities marked `[C]`
- Uses transaction if multi-insert needed (per proto config flag)

#### Acceptance Criteria
- [ ] Test proto with `table { crud: [C,R,U,D] }` produces Create/Update/InsertOrReplace functions in generated output
- [ ] Generated SQL uses correct INSERT ... RETURNING syntax for PostgreSQL
- [ ] Golden test verifies output compiles and contains expected function signatures
- [ ] `conflict_target` proto option properly gates ON CONFLICT clause generation
- [ ] Insert test: creates a row, retrieves it, confirms PK is populated

#### Checkpoint: Phase 1 Complete When ✅ T1.1–T1.4 Pass + Golden Test Updated

---

### Phase 2: Multi-PK & Multi-FK Support (Schema Extensions)

> **Goal**: Extend TableRepo/Column types and writers to support composite primary keys and multiple foreign key references on a single field, which unlock real-world relational patterns.

#### Why Third?
These are schema-level extensions that modify downstream behavior for SchemaWriter, ScannerWriter, AND any new DML writer from Phase 1. Must be done before or alongside Phase 1's update function (which needs multi-PK WHERE clauses).

**Note**: Phase 2 can parallelize with the latter half of Phase 1 (once proto extensions are defined).

#### Tasks

**T2.1: Multi-PK in types.table.go**
- Modify `Table` to support multiple PK columns (`GetPKColumns() []*Column`)
- Update `NewTableFromDescriptor()` to iterate all fields with `pk != PK_UNSPECIFIED`
- Deprecate single-column assumption everywhere

**T2.2: Multi-FK in types.table.go**
- Modify `Table.GetForeignKeys()` to properly return multiple FK references per entity
- Add proto extension field `foreign_keys { to: { entity: "X", fieldnames: ["id1", "id2"] } }` with support for repeated foreign keys per column

**T2.3: Update SchemaWriter for Multi-PK / Multi-FK**
- Generate `tbl.AddPrimary(pkCol1, pkCol2, ...)` (ent supports multi-primary)
- Generate correct FOREIGN KEY syntax with multiple columns in REFERENCES clause
- Verify generated schema.Table is compatible with ent's API

**T2.4: Update ScannerWriter for Multi-PK / Multi-FK**
- Result struct scan handles additional FK subqueries correctly
- `GetPKColumns()` returns all PK column names
- Generated `WHERE ... IN (...)` patterns for composite lookups

#### Acceptance Criteria
- [ ] Proto with `pk: PK_MAN` on multiple columns generates composite primary key in schema
- [ ] Test proto with multi-field FK reference generates correct FOREIGN KEY clause
- [ ] Phase 1's Update function uses multi-PK WHERE clause (cross-phase verification)
- [ ] All golden tests pass including multi-PK / multi-FK test proto

#### Checkpoint: Phase 2 Complete When ✅ T2.1–T2.4 Pass + Cross-phase with Phase 1 verified

---

### Phase 3: Optional Resolution Toggle & oneOf Support

> **Goal**: Give users control over whether FK references are eagerly resolved, and add support for proto3 `oneof` fields as SQL column types.

#### Why Fourth?
Lower-risk additions that don't change core table semantics but improve usability. Optional resolution is a config parameter; oneOf requires type mapping extensions.

#### Tasks

**T3.1: Config parameter injection**
- Extend `Config` struct to include `ResolveReferences bool` (true by default for backward compat)
- In SchemaWriter + ScannerWriter, conditionally generate or skip FK resolution logic based on config
- Users set via protoc parameter: `(sqlmap.go).dialect=postgres.resolve_refs=false`

**T3.2: oneOf support in types/column.go**
- Detect `protoreflect.FieldDescriptor.OneofIndex != nil` in field descriptors
- Generate appropriate SQL type for the first/selected variant of each oneof group
- Handle mixed-type oneofs (string vs int) — pick SQL-compatible type or map to JSON

**T3.3: Generated code for optional resolution**
When `resolve_refs=false`:
- FK columns remain as-is in Result struct (not expanded into embedded entity types)
- No eager load functions generated
- Schema retains simple column definitions without composite JOIN patterns

#### Acceptance Criteria
- [ ] `resolve_refs=false` produces leaner output without FK expansion
- [ ] Test proto with `oneof { string s_val = 1; int32 i_val = 2; }` generates proper column mapping
- [ ] Backward compat: default `resolve_refs=true` unchanged behavior from Phase 0 baseline

#### Checkpoint: Phase 3 Complete When ✅ T3.1–T3.3 Pass + Backward Compat Verified ✓

---

### Phase 4: Naming Improvements & Additional Dialects (3 slices — parallelizable)

> **Goal**: Improve column naming conventions and extend SQL type mapping to CockroachDB, MySQL/MariaDB, MSSQL, and SQLite.

#### Why Last?
Dialect extensions are the most isolated changes — each can evolve independently. Naming improvements are cosmetic but improve real-world usability. Grouping them here because each dialect has naming/formatting differences too.

#### Tasks (Parallelizable)

**T4.1: Naming conventions improvement**
- Add `table_alias` and `column_prefix` options in proto `table` extension
- SchemaWriter uses aliases for generated table/column identifiers
- Column references use alias + column name instead of bare names

**T4.2a: CockroachDB dialect (parallel)**
- Extend dialect enum or mapping in GetSqlType() with `cockroach` key
- Map Go types to PostgreSQL-compatible CRDB syntax (largely same as PG but watch for CRDB-specific types)
- Handle CRDB's unique SQL variations (e.g., UNIQUE NULLS DISTINCT vs PostgreSQL defaults)

**T4.2b: MySQL/MariaDB dialect (parallel)**
- Extend GetSqlType() with more complete MySQL type mapping (TEXT, BLOB, DATETIME, etc.)
- Generate schema.Table entries for both `dialect.Postgres` and `dialect.MySQL` keys per column
- Test ent's MySQL SchemaType behavior

**T4.2c: MSSQL + SQLite dialects (parallel)**
- Map Go/Proto types to MSSQL types (NVARCHAR, NVARCHAR(MAX), DATETIME2, etc.)
- Map Go/Proto types to SQLite affinity types (TEXT, INTEGER, REAL, BLOB)
- Note: ent supports both but with limited type specificity — document limitations

**T4.3: Cross-dialect golden tests**
- For each dialect: same test proto → verify correct SQL Type mapping in generated output
- Schema writer generates `SchemaType: map[string]string{...}` entries for all configured dialects

#### Acceptance Criteria
- [ ] Naming alias option properly scopes table/column identifiers in queries
- [ ] CockroachDB generates valid schema.Table with CRDB-compatible SQL Type keys
- [ ] MySQL maps correctly (TEXT, DECIMAL, TIMESTAMP types etc.)
- [ ] MSSQL uses NVARCHAR, BIT, DATETIME2 etc.
- [ ] SQLite affinity mapping works (TEXT → TEXT, INTEGER → INTEGER, REAL → REAL)
- [ ] All 4 dialects have golden file tests

#### Checkpoint: Phase 4 Complete When ✅ T4.1–T4.3 Pass + All Dialect Golden Tests ✓

---

### Phase 5: Filter API (Query Builder Enhancement)

> **Goal**: Generate chainable filter functions for WHERE clause construction, replacing raw SQL name columns with aliased/table-aware names per the roadmap item "Improve filter to no use column-names".

#### Why Last?
This is a standalone feature that generates Go code — same pattern as SchemaWriter/ScannerWriter but producing *filter builder* helpers. No dependencies on prior phases beyond the existing TableRepo foundation. Can be implemented anytime after Phase 0, but logically completes the "Queries" section of the roadmap.

#### Tasks

**T5.1: Filter type generation per entity**
- Generate functions like `func WithName(value string) func(*sqlmap.Filter)` that build WHERE clauses
- Each column in the entity gets a builder function: `WithName`, `WithAge`, etc.
- Generated as methods on a per-entity type (e.g., `store.Users().Where(user.Name("Alice"))`)

**T5.2: Filter chaining and composition**
- Generated AND/OR combination functions
- Comparison operators: `EQ`, `NEQ`, `LT`, `LTE`, `GT`, `GTE`, `IN`, `LIKE`
- Column names use table alias if aliases are configured (addresses naming improvement from Phase 4)

**T5.3: Integration with Store methods**
- Store query methods accept filter variadic args: `store.Users(ctx, UsersWithName("Alice"), UsersAgeGreaterThan(21))`
- Generated WHERE clause built from filters at runtime

#### Acceptance Criteria
- [ ] Each entity gets per-column filter builder functions
- [ ] Chains of filters produce composed AND clauses
- [ ] Uses aliased column names when aliases are configured
- [ ] Test proto + golden test demonstrates working chainable query

#### Checkpoint: Phase 5 Complete When ✅ T5.1–T5.3 Pass ✓

---

## Task Summary Table

| Phase | Tasks           | Parallel? | Est. Effort | Risk   | Deliverable                             |
|-------|-----------------|-----------|-------------|--------|-----------------------------------------|
| P0    | T0.1-T0.3       | No        | 2-3d        | Low    | Test harness + baseline golden test     |
| P1    | T1.1-T1.4       | Yes (T1.2+T1.3) | 5-7d | Medium | Insert/Update CRUD + conflict handling  |
| P2    | T2.1-T2.4       | Yes (T2.1+T2.2) | 4-5d   | Medium | Multi-PK/Multi-FK table support         |
| P3    | T3.1-T3.3       | No        | 2-3d        | Low    | Config toggle + oneOf support           |
| P4    | T4.1, T4.2a,b,c | Yes      | 5-7d total  | Medium | Naming improvements + 4 new dialects    |
| P5    | T5.1-T5.3       | No        | 3-4d        | Medium | Chainable filter API                    |

**Total estimated effort: ~21-29 developer days across all phases.**

---

## Risk Register

| Risk                                   | Mitigation                                    |
|----------------------------------------|-----------------------------------------------|
| Ent's SchemaType support varies by dialect | Use `ent/dialect` constants; test each at compile time via build tags |
| PostgreSQL vs MySQL type mapping conflicts | Keep dialect maps in separate files; validate against both engines |
| Generated code bloat for large tables  | Add config option to limit per-file output (split by entity) |
| Backward compatibility with generated  | All new proto fields are optional; default behavior unchanged |

---

## Phase Dependencies (Critical Path)

```
P0 ──▶ P2 ───┐
              ├──▶ P1 ──▶ P5
P0 ──▶ P3     │
       ┌──────▶ P4
       │      /
       ▼     /
   (can proceed independently)
```

- **P0 → everything**: Must have tests before feature work
- **P2 → P1**: Multi-PK needed for Update WHERE clauses (P1)
- **P4.1 → filtering naming**: Alias configuration feeds into filter API column names
- **P3 can proceed anytime after P0** (standalone config + oneof)

---

## Decision Logs

### Phase ordering decisions
1. **Tests before features (Phase 0)** — Prevents regression on complex type resolution changes across dialects
2. **Insert/Update after multi-pk foundation partially set up** — Core write operations need table metadata which is stable; multi-PK extends that metadata
3. **Dialect additions parallelized (Phase 4)** — Each dialect is independent SQL type mapping with no shared code paths
4. **Filter API last (Phase 5)** — Depends on naming conventions being stable and table semantics finalized
