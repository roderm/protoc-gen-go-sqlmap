package sqlmap

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// e2eDriverSrc is a Go program written into a throwaway module together with
// the freshly generated .pb.go/.sqlmap.go for testdata/simple.proto. It
// exercises the full path a real consumer would: create the table via
// pkg/migration against a real (in-memory) database, insert a row with a
// plain SQL statement, then read it back through the generated
// SimpleResult.Scan()/GetSimpleColumns(). This is what actually catches
// schema/scanner drift (e.g. mismatched column names) that a text-only
// golden-file comparison cannot.
const e2eDriverSrc = `package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"e2etest/simplepb"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/migration"
	_ "modernc.org/sqlite"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		fail(err)
	}
	defer db.Close()

	m, err := migration.New(db, migration.DialectSQLite)
	if err != nil {
		fail(err)
	}
	if err := m.Create(ctx, simplepb.SimpleTable); err != nil {
		fail(fmt.Errorf("Create: %w", err))
	}

	if _, err := db.ExecContext(ctx,
		"INSERT INTO tbl_simple (simple_id, simple_name) VALUES (?, ?)",
		int64(42), "hello world",
	); err != nil {
		fail(fmt.Errorf("insert: %w", err))
	}

	cols := simplepb.GetSimpleColumns()
	row := db.QueryRowContext(ctx,
		"SELECT "+strings.Join(cols, ", ")+" FROM tbl_simple WHERE simple_id = ?",
		int64(42),
	)

	var result simplepb.SimpleResult
	if err := result.Scan(cols, row); err != nil {
		fail(fmt.Errorf("Scan: %w", err))
	}
	if result.GetId() != 42 {
		fail(fmt.Errorf("Id = %d, want %d", result.GetId(), 42))
	}
	if result.GetName() != "hello world" {
		fail(fmt.Errorf("Name = %q, want %q", result.GetName(), "hello world"))
	}
	fmt.Println("OK")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
`

// TestE2E_MigrationCreateAndScan generates testdata/simple.proto through the
// real pipeline (protoc-gen-go for the base message + this generator for the
// schema/scanner), assembles it into a throwaway Go module, and runs a
// driver program that creates the table via pkg/migration, inserts a row,
// and reads it back through the generated Scan(). It requires network/module
// cache access to resolve modernc.org/sqlite, so it's opt-in.
func TestE2E_MigrationCreateAndScan(t *testing.T) {
	if os.Getenv("SQLMAP_E2E") == "" {
		t.Skip("set SQLMAP_E2E=1 to run (needs protoc, protoc-gen-go, and network/module-cache access for modernc.org/sqlite)")
	}
	if _, err := exec.LookPath("protoc"); err != nil {
		t.Skip("protoc not found in PATH")
	}
	if _, err := exec.LookPath("protoc-gen-go"); err != nil {
		t.Skip("protoc-gen-go not found in PATH")
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}

	// 1. Generate the base message type (Simple struct) via the real
	// protoc-gen-go binary, into its own temp dir.
	pbOutDir := t.TempDir()
	cmd := exec.Command("protoc",
		"-I", filepath.Join(repoRoot, "proto"),
		"-I", testdata,
		"--go_out=paths=source_relative,"+
			"Msqlmap/v1/sqlmap.proto=github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1,"+
			"Mschema/v1/schema.proto=github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1:"+pbOutDir,
		filepath.Join(testdata, "simple.proto"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc --go_out: %v\n%s", err, out)
	}
	basePB, err := os.ReadFile(filepath.Join(pbOutDir, "simple.pb.go"))
	if err != nil {
		t.Fatal(err)
	}

	// 2. Generate the schema/scanner code via our own generator, in-process.
	files := generate(t, "simple.proto")
	schemaGo, ok := files["simple.sqlmap.go"]
	if !ok {
		t.Fatalf("no simple.sqlmap.go in generated output, got files: %v", fileNames(files))
	}

	// 3. Assemble a throwaway module containing both generated files, the
	// driver program above, and a replace directive pointing back at this
	// repo checkout so pkg/migration resolves to the code under test.
	modRoot := t.TempDir()
	pkgDir := filepath.Join(modRoot, "simplepb")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "simple.pb.go"), basePB, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "simple.sqlmap.go"), []byte(schemaGo), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modRoot, "main.go"), []byte(e2eDriverSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(modRoot, "go.mod"), []byte("module e2etest\n\ngo 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	run := func(name string, args ...string) string {
		t.Helper()
		c := exec.Command(name, args...)
		c.Dir = modRoot
		out, err := c.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, out)
		}
		return string(out)
	}
	run("go", "mod", "edit",
		"-require=github.com/roderm/protoc-gen-go-sqlmap@v0.0.0",
		"-replace=github.com/roderm/protoc-gen-go-sqlmap="+repoRoot,
	)
	run("go", "mod", "tidy")

	out := run("go", "run", ".")
	if out != "OK\n" {
		t.Fatalf("driver program did not report success, got: %q", out)
	}
}
