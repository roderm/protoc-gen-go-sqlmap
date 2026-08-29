package query

import (
	"testing"

	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestMask_NilSelectsEverything(t *testing.T) {
	var m *Mask
	if !m.Has("anything") {
		t.Error("a nil mask must select everything")
	}
	if m.Sub("anything") != nil {
		t.Error("a nil mask's sub-mask must also select everything")
	}
	if NewMask() != nil {
		t.Error("NewMask() with no paths must select everything")
	}
	if FromFieldMask(nil) != nil {
		t.Error("FromFieldMask(nil) must select everything")
	}
	if FromFieldMask(&fieldmaskpb.FieldMask{}) != nil {
		t.Error("an empty FieldMask must select everything")
	}
}

func TestMask_Leaf(t *testing.T) {
	m := NewMask("name")
	if !m.Has("name") {
		t.Error(`Has("name") = false, want true`)
	}
	if m.Has("other") {
		t.Error(`Has("other") = true, want false`)
	}
	// A leaf path selects the whole subtree below it.
	if m.Sub("name") != nil {
		t.Error(`Sub("name") must select everything below a leaf path`)
	}
}

func TestMask_Nested(t *testing.T) {
	m := NewMask("books.publisher.name")
	if !m.Has("books") {
		t.Fatal(`Has("books") = false, want true`)
	}
	if m.Has("name") {
		t.Error(`Has("name") = true at the root, want false`)
	}
	books := m.Sub("books")
	if books == nil {
		t.Fatal(`Sub("books") must be narrower than "everything"`)
	}
	if !books.Has("publisher") || books.Has("title") {
		t.Error("the books sub-mask should select only publisher")
	}
	if pub := books.Sub("publisher"); pub == nil || !pub.Has("name") || pub.Has("city") {
		t.Error("the publisher sub-mask should select only name")
	}
}

// A shorter path selects everything below it, and must win regardless of the
// order it appears in relative to a longer path through the same node.
func TestMask_BroadPathWinsInEitherOrder(t *testing.T) {
	for _, paths := range [][]string{
		{"books", "books.title"},
		{"books.title", "books"},
	} {
		m := NewMask(paths...)
		if !m.Has("books") {
			t.Fatalf("%v: Has(\"books\") = false, want true", paths)
		}
		if sub := m.Sub("books"); sub != nil {
			t.Errorf("%v: Sub(\"books\") = %v, want nil (selects everything)", paths, sub)
		}
	}
}

func TestMask_SiblingPaths(t *testing.T) {
	m := NewMask("books.title", "books.isbn", "name")
	if !m.Has("name") || !m.Has("books") {
		t.Fatal("both top-level paths should be selected")
	}
	books := m.Sub("books")
	if books == nil {
		t.Fatal("books should have a narrower sub-mask")
	}
	if !books.Has("title") || !books.Has("isbn") || books.Has("author") {
		t.Error("the books sub-mask should select exactly title and isbn")
	}
}

func TestIn(t *testing.T) {
	if _, ok := In("id", nil); ok {
		t.Error("In with no values must report false, since `IN ()` is invalid SQL")
	}
	cond, ok := In("id", []any{int64(1), int64(2)})
	if !ok {
		t.Fatal("In with values must report true")
	}
	if cond.SQL != "id IN (?, ?)" {
		t.Errorf("SQL = %q, want %q", cond.SQL, "id IN (?, ?)")
	}
	if len(cond.Args) != 2 {
		t.Errorf("got %d args, want 2", len(cond.Args))
	}
}

func TestConn_RenumberPlaceholders(t *testing.T) {
	// PostgreSQL placeholders are numbered across the whole statement, so the
	// second condition must continue where the first left off.
	pg := Conn{Dialect: Postgres}
	if got := pg.renumber("a = ? AND b = ?", 0); got != "a = $1 AND b = $2" {
		t.Errorf("got %q, want %q", got, "a = $1 AND b = $2")
	}
	if got := pg.renumber("c = ?", 2); got != "c = $3" {
		t.Errorf("got %q, want %q", got, "c = $3")
	}
	// Everything else keeps positional placeholders untouched.
	for _, d := range []Dialect{MySQL, SQLite} {
		if got := (Conn{Dialect: d}).renumber("a = ?", 0); got != "a = ?" {
			t.Errorf("%s: got %q, want %q", d, got, "a = ?")
		}
	}
}

// Key exists so a value read back from a driver and one read off a generated
// getter compare equal as map keys, whatever width or byte-ness the driver
// chose to hand back.
func TestKey_NormalizesAcrossDriverTypes(t *testing.T) {
	if Key(int32(7)) != Key(int64(7)) {
		t.Error("int32 and int64 join keys must compare equal")
	}
	if Key([]byte("abc")) != Key("abc") {
		t.Error("[]byte and string join keys must compare equal")
	}
	if Key(nil) != nil {
		t.Error("a nil value must stay nil so callers can skip it")
	}
}
