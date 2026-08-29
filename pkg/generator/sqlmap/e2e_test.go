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

// relationDriverSrc asserts the properties only a live database can confirm:
// that a nullable column really accepts NULL, that a NOT NULL column rejects
// it, and that the generated foreign key actually applies ON DELETE SET NULL.
const relationDriverSrc = `package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"e2etest/relationpb"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/migration"
)

func run(ctx context.Context, db *sql.DB) error {
	m, err := migration.New(db, migration.DialectPostgres)
	if err != nil {
		return err
	}
	if err := m.Create(ctx, relationpb.AuthorTable, relationpb.BookTable); err != nil {
		return fmt.Errorf("Create: %w", err)
	}

	var authorID int64
	if err := db.QueryRowContext(ctx,
		"INSERT INTO tbl_author (author_name) VALUES ($1) RETURNING author_id", "Ursula",
	).Scan(&authorID); err != nil {
		return fmt.Errorf("insert author: %w", err)
	}
	var bookID int64
	if err := db.QueryRowContext(ctx,
		"INSERT INTO tbl_book (book_title, author_id) VALUES ($1, $2) RETURNING book_id", "Earthsea", authorID,
	).Scan(&bookID); err != nil {
		return fmt.Errorf("insert book: %w", err)
	}

	// author_id has proto presence and is not a PK, so it must be nullable.
	if _, err := db.ExecContext(ctx,
		"INSERT INTO tbl_book (book_title, author_id) VALUES ($1, NULL)", "Orphan",
	); err != nil {
		return fmt.Errorf("insert NULL author_id, so the column was generated NOT NULL: %w", err)
	}
	// book_id is a PK: it must have been generated NOT NULL.
	if _, err := db.ExecContext(ctx, "INSERT INTO tbl_book (book_id) VALUES (NULL)"); err == nil {
		return fmt.Errorf("inserting NULL into the primary key succeeded, so it was generated nullable")
	}
	// Deleting the author must NULL the referencing column rather than fail,
	// which only happens if ON DELETE SET NULL made it into the DDL.
	if _, err := db.ExecContext(ctx, "DELETE FROM tbl_author WHERE author_id = $1", authorID); err != nil {
		return fmt.Errorf("delete author (ON DELETE SET NULL not applied?): %w", err)
	}

	cols := relationpb.GetBookColumns()
	row := db.QueryRowContext(ctx,
		"SELECT "+strings.Join(cols, ", ")+" FROM tbl_book WHERE book_id = $1", bookID)
	var result relationpb.BookResult
	if err := result.Scan(cols, row); err != nil {
		return fmt.Errorf("Scan: %w", err)
	}
	if result.GetTitle() != "Earthsea" {
		return fmt.Errorf("Title = %q, want %q", result.GetTitle(), "Earthsea")
	}
	if result.AuthorId != nil {
		return fmt.Errorf("AuthorId = %d, want NULL after the author was deleted", *result.AuthorId)
	}
	return nil
}
`

// eagerDriverSrc exercises the query writer: that a FieldMask restricts the
// columns actually selected, that it decides which relations are loaded, and
// that nested paths load two levels deep in both relation directions
// (has-many Author->Books, belongs-to Book->Publisher).
const eagerDriverSrc = `package main

import (
	"context"
	"database/sql"
	"fmt"

	"e2etest/eagerpb"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/migration"
	"github.com/roderm/protoc-gen-go-sqlmap/pkg/query"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func run(ctx context.Context, db *sql.DB) error {
	m, err := migration.New(db, migration.DialectPostgres)
	if err != nil {
		return err
	}
	if err := m.Create(ctx, eagerpb.PublisherTable, eagerpb.AuthorTable, eagerpb.BookTable); err != nil {
		return fmt.Errorf("Create: %w", err)
	}

	var pubID, authorID int64
	if err := db.QueryRowContext(ctx,
		"INSERT INTO tbl_publisher (publisher_name) VALUES ($1) RETURNING publisher_id", "Gollancz",
	).Scan(&pubID); err != nil {
		return err
	}
	if err := db.QueryRowContext(ctx,
		"INSERT INTO tbl_author (author_name) VALUES ($1) RETURNING author_id", "Ursula",
	).Scan(&authorID); err != nil {
		return err
	}
	for _, title := range []string{"Earthsea", "Lathe"} {
		if _, err := db.ExecContext(ctx,
			"INSERT INTO tbl_book (book_title, author_id, publisher_id) VALUES ($1, $2, $3)",
			title, authorID, pubID); err != nil {
			return err
		}
	}
	// A second author with no books, to prove eager loading does not invent one.
	if _, err := db.ExecContext(ctx, "INSERT INTO tbl_author (author_name) VALUES ($1)", "Nobody"); err != nil {
		return err
	}

	conn := query.Conn{DB: db, Dialect: query.Postgres}

	// Two levels of nesting, across both relation directions.
	authors, err := eagerpb.LoadAuthor(ctx, conn,
		&fieldmaskpb.FieldMask{Paths: []string{"name", "books.title", "books.publisher.name"}})
	if err != nil {
		return fmt.Errorf("LoadAuthor: %w", err)
	}
	if len(authors) != 2 {
		return fmt.Errorf("got %d authors, want 2", len(authors))
	}
	byName := map[string]*eagerpb.Author{}
	for _, a := range authors {
		byName[a.GetName()] = a
	}
	ursula, ok := byName["Ursula"]
	if !ok {
		return fmt.Errorf("author Ursula missing; got %v", byName)
	}
	if len(ursula.GetBooks()) != 2 {
		return fmt.Errorf("Ursula has %d books, want 2", len(ursula.GetBooks()))
	}
	if got := byName["Nobody"].GetBooks(); len(got) != 0 {
		return fmt.Errorf("Nobody has %d books, want 0", len(got))
	}
	for _, b := range ursula.GetBooks() {
		if b.GetTitle() == "" {
			return fmt.Errorf("book title was masked in but came back empty")
		}
		if b.GetPublisher().GetName() != "Gollancz" {
			return fmt.Errorf("nested publisher name = %q, want %q", b.GetPublisher().GetName(), "Gollancz")
		}
	}

	// A mask that does not name the relation must not load it, and must not
	// select the columns it left out either.
	authors, err = eagerpb.LoadAuthor(ctx, conn, &fieldmaskpb.FieldMask{Paths: []string{"name"}})
	if err != nil {
		return fmt.Errorf("LoadAuthor(name): %w", err)
	}
	for _, a := range authors {
		if len(a.GetBooks()) != 0 {
			return fmt.Errorf("books were loaded despite not being in the mask")
		}
	}

	// Conversely, masking only the relation must leave the scalar unselected.
	authors, err = eagerpb.LoadAuthor(ctx, conn, &fieldmaskpb.FieldMask{Paths: []string{"books.title"}})
	if err != nil {
		return fmt.Errorf("LoadAuthor(books.title): %w", err)
	}
	for _, a := range authors {
		if a.GetName() != "" {
			return fmt.Errorf("author_name = %q was selected despite not being in the mask", a.GetName())
		}
	}

	// A nil mask selects every column but loads no relations: eager loading is
	// opt-in, so that "no mask" cannot walk the whole object graph.
	authors, err = eagerpb.LoadAuthor(ctx, conn, nil)
	if err != nil {
		return fmt.Errorf("LoadAuthor(nil): %w", err)
	}
	for _, a := range authors {
		if a.GetName() == "" {
			return fmt.Errorf("a nil mask should select every column, but author_name was empty")
		}
		if len(a.GetBooks()) != 0 {
			return fmt.Errorf("a nil mask loaded %d books; relations must be opt-in", len(a.GetBooks()))
		}
	}

	// A leaf path selects the whole relation, but must not cascade into the
	// relations below it.
	authors, err = eagerpb.LoadAuthor(ctx, conn, &fieldmaskpb.FieldMask{Paths: []string{"books"}})
	if err != nil {
		return fmt.Errorf("LoadAuthor(books): %w", err)
	}
	for _, a := range authors {
		for _, b := range a.GetBooks() {
			if b.GetTitle() == "" {
				return fmt.Errorf("a leaf relation path should select all of its columns")
			}
			if b.GetPublisher() != nil {
				return fmt.Errorf("a leaf relation path must not cascade into nested relations")
			}
		}
	}
	return nil
}
`

// mainSrc wraps a driver's run() with connection setup, so each driver only
// contains the assertions that make it distinct.
const mainSrc = `package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	db, err := sql.Open("postgres", os.Getenv("SQLMAP_DSN"))
	if err != nil {
		fail(err)
	}
	defer db.Close()

	deadline := time.Now().Add(60 * time.Second)
	for {
		if err := db.PingContext(ctx); err == nil {
			break
		} else if time.Now().After(deadline) {
			fail(fmt.Errorf("database never became reachable: %w", err))
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err := run(ctx, db); err != nil {
		fail(err)
	}
	fmt.Println("OK")
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
			// One "addr:port" mapping per line (IPv4 and IPv6).
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

// runE2E generates protoFile through the real pipeline (protoc-gen-go for the
// base messages plus this generator for the schema/scanner/query code),
// assembles it into a throwaway Go module alongside driverSrc, and runs it
// against a freshly started PostgreSQL.
//
// The SQL driver is required by *that* module, not by the repo under test, so
// the generator keeps its two-dependency footprint.
func runE2E(t *testing.T, protoFile, pkgName, driverSrc string) {
	t.Helper()
	if os.Getenv("SQLMAP_E2E") == "" {
		t.Skip("set SQLMAP_E2E=1 to run (needs docker, protoc, protoc-gen-go, and module-cache access)")
	}
	for _, bin := range []string{"protoc", "protoc-gen-go"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skipf("%s not found in PATH", bin)
		}
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}
	base := strings.TrimSuffix(protoFile, ".proto")
	dsn := startPostgres(t)

	// 1. Base message types, via the real protoc-gen-go binary.
	pbOutDir := t.TempDir()
	cmd := exec.Command("protoc",
		"-I", filepath.Join(repoRoot, "proto"),
		"-I", testdata,
		"--go_out=paths=source_relative,"+
			"Msqlmap/v1/sqlmap.proto=github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/sqlmap/v1,"+
			"Mschema/v1/schema.proto=github.com/roderm/protoc-gen-go-sqlmap/pkg/generated/schema/v1:"+pbOutDir,
		filepath.Join(testdata, protoFile),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("protoc --go_out: %v\n%s", err, out)
	}
	basePB, err := os.ReadFile(filepath.Join(pbOutDir, base+".pb.go"))
	if err != nil {
		t.Fatal(err)
	}

	// 2. Schema/scanner/query code, via our own generator, in-process.
	files := generate(t, protoFile)
	generated, ok := files[base+".sqlmap.go"]
	if !ok {
		t.Fatalf("no %s.sqlmap.go in generated output, got files: %v", base, fileNames(files))
	}

	// 3. A throwaway module holding both generated files and the driver, with
	// a replace directive pointing back at this checkout so pkg/migration and
	// pkg/query resolve to the code under test.
	modRoot := t.TempDir()
	pkgDir := filepath.Join(modRoot, pkgName)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string][]byte{
		filepath.Join(pkgDir, base+".pb.go"):     basePB,
		filepath.Join(pkgDir, base+".sqlmap.go"): []byte(generated),
		filepath.Join(modRoot, "main.go"):        []byte(mainSrc),
		filepath.Join(modRoot, "driver.go"):      []byte(driverSrc),
		filepath.Join(modRoot, "go.mod"):         []byte("module e2etest\n\ngo 1.27\n"),
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

// TestE2E_MigrationCreateAndScan covers the schema writer and pkg/migration:
// nullability and ON DELETE SET NULL against a live database.
func TestE2E_MigrationCreateAndScan(t *testing.T) {
	runE2E(t, "relation.proto", "relationpb", relationDriverSrc)
}

// TestE2E_EagerLoading covers the query writer: FieldMask-driven column
// selection and eager loading across both relation directions.
func TestE2E_EagerLoading(t *testing.T) {
	runE2E(t, "eager.proto", "eagerpb", eagerDriverSrc)
}
