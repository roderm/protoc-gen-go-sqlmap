# DESIGN: subtype tables (`Identity` is a `Person` or an `Organisation`)

Status: proposal. Nothing here is implemented yet.

## The problem

The CRM service models an identity that is *exactly one of* a person or an organisation:

```protobuf
message Identity {
  option (sqlmap.v1.table) = {name: "identity"};
  string id = 1 [...];
  oneof type {
    Person person = 5;
    Organisation organisation = 6;
  }
}
```

`Person` and `Organisation` are their own tables whose primary key is a foreign key back to `identity.id` — joined-table inheritance. The proto says "exactly one of", but the database says nothing at all: today you can create an identity with a `person` row *and* an `organisation` row, or with neither, and nothing complains.

The generator also cannot navigate this relation. A singular relation is assumed to be a belongs-to, where the row holds the key; here the key lives on the subtype row. So `Identity.person` is neither constrained nor loadable.

## What SQL can and cannot enforce

Two separate guarantees, and they are not equally achievable:

| Guarantee | Declarative SQL? |
|---|---|
| **At most one** subtype row per supertype row | Yes — the pattern below |
| **At least one** subtype row per supertype row | No |

"At least one" is a chicken-and-egg problem: the supertype row must exist before the subtype row can reference it, so there is always an instant where it has no subtype. Enforcing it needs `DEFERRABLE INITIALLY DEFERRED` constraint triggers that fire at commit — non-declarative, PostgreSQL-specific, and beyond what a schema generator should emit. **The proposal enforces at-most-one and leaves at-least-one to the application.** That is the normal trade-off for this pattern.

## The pattern: discriminator + composite foreign key

Add a discriminator column to the supertype saying which subtype a row is. Give each subtype table the same column, pinned by a CHECK to its own single value. Then make the subtype's foreign key composite, over `(key, discriminator)` rather than just `key`.

An identity has exactly one `kind` value, and a `person` row can only reference an identity whose `kind` is `'person'` — so only one subtype table can ever hold a row for a given identity.

```sql
CREATE TABLE identity (
  id   VARCHAR(64) NOT NULL PRIMARY KEY,
  kind VARCHAR(32) NOT NULL,
  CONSTRAINT identity_kind_check CHECK (kind IN ('person','organisation')),
  CONSTRAINT identity_id_kind_key UNIQUE (id, kind)      -- the FK target
);

CREATE TABLE person (
  id         VARCHAR(64) NOT NULL PRIMARY KEY,
  kind       VARCHAR(32) NOT NULL DEFAULT 'person',
  given_name VARCHAR(255),
  CONSTRAINT person_kind_check CHECK (kind = 'person'),
  CONSTRAINT person_identity_fk
    FOREIGN KEY (id, kind) REFERENCES identity (id, kind) ON DELETE CASCADE
);
-- organisation is the same shape with 'organisation'
```

The `UNIQUE (id, kind)` is redundant data-wise (`id` is already the primary key) but required: a foreign key must reference a uniquely-constrained column set.

### Verified against PostgreSQL 16

Each of these is rejected by the database:

| Attempt | Rejected by |
|---|---|
| `organisation` row for an identity whose kind is `'person'` | `organisation_identity_fk` |
| Flipping `identity.kind` while a `person` row exists | `person_identity_fk` |
| An identity with an unknown kind (`'robot'`) | `identity_kind_check` |
| A `person` row with `kind = 'organisation'` forced in directly | `person_kind_check` |

`DELETE FROM identity` cascades the subtype row away, as expected.

## Proposed proto surface

Two declarations: the supertype names the discriminator column, each subtype names its parent.

### Supertype — an option on the `oneof`

`oneof` carries options via `google.protobuf.OneofOptions`, which is exactly the right place: the oneof already enumerates the subtypes and already means "exactly one of".

```protobuf
extend google.protobuf.OneofOptions {
  optional Subtypes subtypes = 800140;
}

// Subtypes marks a oneof of message fields as a joined-table subtype
// hierarchy rooted at this message.
message Subtypes {
  // Column added to THIS table recording which subtype each row is.
  required string discriminator = 1;
  // SQL type of that column per dialect. Defaults to VARCHAR(32).
  map<string, string> type = 2;
}
```

### Subtype — a field on `Table`

```protobuf
message Table {
  required string name = 1;
  repeated ForeignKeyDefinition foreign_keys = 2;
  optional SubtypeOf subtype_of = 3;              // new
}

message SubtypeOf {
  // The supertype message.
  required string entity = 1;
  // Value stored in the supertype's discriminator for this subtype.
  // Defaults to the oneof field's name on the supertype.
  optional string value = 2;
  // Columns on THIS table carrying the supertype's key.
  // Defaults to this table's primary key.
  repeated string fieldnames = 3;
  optional schema.v1.OnDelete on_delete = 4;
}
```

### How it reads in the CRM protos

```protobuf
message Identity {
  option (sqlmap.v1.table) = {name: "identity"};
  string id = 1 [(sqlmap.v1.col) = {fieldname: "id", pk: PK_MAN, ...}];

  oneof type {
    option (sqlmap.v1.subtypes) = {discriminator: "kind"};
    Person person = 5;
    Organisation organisation = 6;
  }
}

message Person {
  option (sqlmap.v1.table) = {
    name: "person"
    subtype_of: {entity: "Identity", on_delete: ON_DELETE_CASCADE}
  };
  // The explicit foreign_key on `id` goes away -- subtype_of implies it,
  // and as a composite key the old single-column one would be wrong.
  string id = 1 [(sqlmap.v1.col) = {fieldname: "id", pk: PK_MAN, ...}];
  string given_name = 2 [...];
}
```

Both sides declare, deliberately. The generator *could* synthesize the supertype's discriminator purely from the subtypes that declare `subtype_of`, but the set of subtypes it can see depends on which files happen to be compiled in one `protoc` invocation — so generating `person.proto` alone would emit a narrower `CHECK` than generating it alongside `organisation.proto`. A schema that silently changes with the compilation unit is not acceptable; the supertype declaring its own column keeps it deterministic.

## What the generator has to emit

**Supertype table**
- A discriminator column, `NOT NULL`.
- `CHECK (kind IN (<every declared subtype value>))`.
- `UNIQUE (<pk columns>, kind)` so the composite FK has a target.

**Each subtype table**
- The same discriminator column, `NOT NULL DEFAULT '<value>'`.
- `CHECK (kind = '<value>')`.
- A composite foreign key `(<fieldnames>, kind) -> supertype(<pk>, kind)`.

### Required `schema.v1` additions

The intermediate schema cannot currently express any of this:

```protobuf
message SchemaCheck {
  optional string name = 1;
  optional string expr = 2;
}

message SchemaIndex {
  optional string name = 1;
  repeated SchemaColumn columns = 2;
  optional bool unique = 3;
}

message SchemaColumn {
  ...
  optional string default = 5;      // raw SQL default expression
}

message SchemaTable {
  ...
  repeated SchemaCheck checks = 4;
  repeated SchemaIndex indexes = 5;
}
```

These are worth having regardless — they are the same primitives the outstanding "no indexes, no unique constraints, no defaults" gap needs, so this design pays for that too.

`ariga/atlas` already has everything on the receiving side: `schema.NewCheck()` / `Table.AddChecks()`, `schema.NewUniqueIndex()`, and composite foreign keys, which `toSchema` builds already.

## Why this is also the fix for eager loading

The discriminator is not only a constraint — it is what makes `Identity -> Person | Organisation` loadable at all.

Without it, resolving the subtype means querying *every* subtype table for every identity id and seeing which one answers: N queries per batch, most returning nothing. With it, the loader reads `kind` alongside the identity rows, partitions the batch by discriminator value, and issues **one query per subtype actually present** — the same batched shape the existing has-many loading uses.

So `Identity.person` and `Identity.organisation` become ordinary mask paths:

```go
crmv1.LoadIdentity(ctx, conn, &fieldmaskpb.FieldMask{
    Paths: []string{"person.given_name", "organisation.name"},
})
```

A oneof needs one extra step over a normal relation: assigning the loaded message into the generated wrapper type (`&Identity_Person{Person: row}`) rather than a plain pointer field.

## Alternatives considered

**Single-table inheritance** — one `identity` table with every subtype's columns, all nullable, plus a CHECK that the right ones are non-null per kind. No joins and no composite keys, but every subtype column loses its `NOT NULL`, the table grows with each new subtype, and it does not match the existing schema. Reasonable when subtypes differ by one or two columns; not here.

**No supertype table** — drop `identity`, keep `person` and `organisation` standalone. Then `email` and `membership` have nothing single to point at, which is the whole reason `identity` exists.

**Exclusive arc on the child** — give `email` both `person_id` and `organisation_id`, nullable, with a CHECK that exactly one is set. This solves a *different* problem (a child referencing one of several parents) and scales badly: every referencing table grows a column per subtype, and none of them can be a single foreign-key target.

**`CHECK` with a subquery** — `CHECK (NOT EXISTS (SELECT ...))` would express at-most-one directly, but no major database allows subqueries in a CHECK.

**Triggers** — can enforce both directions, including at-least-one, but they are procedural, differ per dialect, and are not something this generator should be emitting.

## Open questions

1. **Discriminator values.** Defaulting to the oneof field name (`person`, `organisation`) is readable in the database and needs no extra declaration. Renaming a proto field would then be a schema change — acceptable, or should `value` be mandatory?
2. **At-least-one.** Left to the application. Worth an opt-in `DEFERRABLE` trigger for PostgreSQL later, or better left alone?
3. **Nested hierarchies.** A subtype that is itself a supertype is not addressed; the design does not forbid it, but nothing validates the chain either.
4. **`PK_AUTO` supertypes.** With a generated key the application must read the supertype's id back before inserting the subtype row. Fine, but it means subtype insertion is inherently two statements and wants a transaction — relevant once insert generation exists.
