package sqlmap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// e2eDriverSrc is a Go program written into a throwaway module together with
// the freshly generated .pb.go/.sqlmap.go for testdata/relation.proto. It
// exercises the full path a real consumer would: create the tables via
// pkg/migration against a real PostgreSQL, then assert the properties that
// only a live database can confirm -- that a nullable column really accepts
// NULL, that a NOT NULL column rejects it, and that the generated foreign key
// actually applies ON DELETE SET NULL. This is what catches schema/scanner
// drift that a text-only golden-file comparison cannot.
//
// The SQL driver is required by *this* module, not by the repo under test, so
// the generator keeps its two-dependency footprint.
const e2eDriverSrc = `package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"e2etest/relationpb"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/migration"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("postgres", os.Getenv("SQLMAP_DSN"))
	if err != nil {
		fail(err)
	}
	defer db.Close()
	waitReady(ctx, db)

	m, err := migration.New(db, migration.DialectPostgres)
	if err != nil {
		fail(err)
	}
	if err := m.Create(ctx, relationpb.AuthorTable, relationpb.BookTable); err != nil {
		fail(fmt.Errorf("Create: %w", err))
	}

	var authorID int64
	if err := db.QueryRowContext(ctx,
		"INSERT INTO tbl_author (author_name) VALUES ($1) RETURNING author_id", "Ursula",
	).Scan(&authorID); err != nil {
		fail(fmt.Errorf("insert author: %w", err))
	}

	var bookID int64
	if err := db.QueryRowContext(ctx,
		"INSERT INTO tbl_book (book_title, author_id) VALUES ($1, $2) RETURNING book_id", "A Wizard of Earthsea", authorID,
	).Scan(&bookID); err != nil {
		fail(fmt.Errorf("insert book: %w", err))
	}

	// author_id has proto presence and is not a PK, so it must be nullable.
	if _, err := db.ExecContext(ctx,
		"INSERT INTO tbl_book (book_title, author_id) VALUES ($1, NULL)", "Orphan",
	); err != nil {
		fail(fmt.Errorf("insert book with NULL author_id, so the column was generated NOT NULL: %w", err))
	}

	// book_id is a PK: it must have been generated NOT NULL.
	if _, err := db.ExecContext(ctx, "INSERT INTO tbl_book (book_id) VALUES (NULL)"); err == nil {
		fail(fmt.Errorf("inserting NULL into the primary key succeeded, so the column was generated nullable"))
	}

	// Deleting the author must NULL out the referencing column rather than
	// fail, which only happens if ON DELETE SET NULL made it into the DDL.
	if _, err := db.ExecContext(ctx, "DELETE FROM tbl_author WHERE author_id = $1", authorID); err != nil {
		fail(fmt.Errorf("delete author (ON DELETE SET NULL not applied?): %w", err))
	}

	cols := relationpb.GetBookColumns()
	row := db.QueryRowContext(ctx,
		"SELECT "+strings.Join(cols, ", ")+" FROM tbl_book WHERE book_id = $1", bookID,
	)
	var result relationpb.BookResult
	if err := result.Scan(cols, row); err != nil {
		fail(fmt.Errorf("Scan: %w", err))
	}
	if result.GetTitle() != "A Wizard of Earthsea" {
		fail(fmt.Errorf("Title = %q, want %q", result.GetTitle(), "A Wizard of Earthsea"))
	}
	if result.AuthorId != nil {
		fail(fmt.Errorf("AuthorId = %d, want NULL after the author was deleted", *result.AuthorId))
	}
	fmt.Println("OK")
}

func waitReady(ctx context.Context, db *sql.DB) {
	deadline := time.Now().Add(60 * time.Second)
	for {
		if err := db.PingContext(ctx); err == nil {
			return
		} else if time.Now().After(deadline) {
			fail(fmt.Errorf("database never became reachable: %w", err))
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
`

const (
	pgImage    = "postgres:16-alpine"
	pgUser     = "sqlmap"
	pgPassword = "sqlmap"
	pgDatabase = "sqlmap"
)

// startPostgres runs a throwaway PostgreSQL container and returns a DSN for
// it. The container is removed when the test finishes.
func startPostgres(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found in PATH")
	}

	// Output(), not CombinedOutput(): the container id goes to stdout while
	// image-pull progress goes to stderr, and mixing them corrupts the id.
	out, err := exec.Command("docker", "run", "-d", "--rm",
		"-e", "POSTGRES_USER="+pgUser,
		"-e", "POSTGRES_PASSWORD="+pgPassword,
		"-e", "POSTGRES_DB="+pgDatabase,
		"-P", pgImage,
	).Output()
	if err != nil {
		t.Skipf("could not start %s: %v", pgImage, err)
	}
	id := strings.TrimSpace(string(out))
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", id).Run() })

	port, err := hostPort(id)
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("postgres://%s:%s@127.0.0.1:%s/%s?sslmode=disable",
		pgUser, pgPassword, port, pgDatabase)
}

// hostPort resolves the host port docker mapped the container's 5432 to.
func hostPort(id string) (string, error) {
	deadline := time.Now().Add(30 * time.Second)
	for {
		out, err := exec.Command("docker", "port", id, "5432/tcp").Output()
		if err == nil {
			// Output is one "addr:port" mapping per line (IPv4 and IPv6).
			for line := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
				if i := strings.LastIndex(line, ":"); i >= 0 {
					return strings.TrimSpace(line[i+1:]), nil
				}
			}
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("container %s never published port 5432: %v", id, err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// TestE2E_MigrationCreateAndScan generates testdata/relation.proto through the
// real pipeline (protoc-gen-go for the base messages + this generator for the
// schema/scanner), assembles it into a throwaway Go module, and runs a driver
// program that creates the tables via pkg/migration against a dockerized
// PostgreSQL and asserts nullability and ON DELETE SET NULL behaviour. It
// needs docker plus network/module-cache access, so it's opt-in.
func TestE2E_MigrationCreateAndScan(t *testing.T) {
	if os.Getenv("SQLMAP_E2E") == "" {
		t.Skip("set SQLMAP_E2E=1 to run (needs docker, protoc, protoc-gen-go, and module-cache access)")
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

	dsn := startPostgres(t)

	// 1. Generate the base message types via the real protoc-gen-go binary.
	pbOutDir := t.TempDir()
	cmd := exec.Command("protoc",
		"-I", filepath.Join(repoRoot, "proto"),
		"-I", testdata,
		"--go_out=paths=source_relative,"+
			"Msqlmap/v1/sqlmap.proto=github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1,"+
			"Mschema/v1/schema.proto=github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1:"+pbOutDir,
		filepath.Join(testdata, "relation.proto"),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc --go_out: %v\n%s", err, out)
	}
	basePB, err := os.ReadFile(filepath.Join(pbOutDir, "relation.pb.go"))
	if err != nil {
		t.Fatal(err)
	}

	// 2. Generate the schema/scanner code via our own generator, in-process.
	files := generate(t, "relation.proto")
	schemaGo, ok := files["relation.sqlmap.go"]
	if !ok {
		t.Fatalf("no relation.sqlmap.go in generated output, got files: %v", fileNames(files))
	}

	// 3. Assemble a throwaway module containing both generated files, the
	// driver program above, and a replace directive pointing back at this
	// repo checkout so pkg/migration resolves to the code under test.
	modRoot := t.TempDir()
	pkgDir := filepath.Join(modRoot, "relationpb")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string][]byte{
		filepath.Join(pkgDir, "relation.pb.go"):     basePB,
		filepath.Join(pkgDir, "relation.sqlmap.go"): []byte(schemaGo),
		filepath.Join(modRoot, "main.go"):           []byte(e2eDriverSrc),
		filepath.Join(modRoot, "go.mod"):            []byte("module e2etest\n\ngo 1.27\n"),
	} {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	run := func(name string, args ...string) string {
		t.Helper()
		c := exec.Command(name, args...)
		c.Dir = modRoot
		c.Env = append(os.Environ(), "SQLMAP_DSN="+dsn)
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

	if out := run("go", "run", "."); out != "OK\n" {
		t.Fatalf("driver program did not report success, got: %q", out)
	}
}
