package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var protoPath = "test/proto"
var fdescRE = regexp.MustCompile(`(?ms)^var fileDescriptor.*}`)

var protoExecutable = "protoc" // "protoc-min-version", "--version=3.0.0"

// Set --regenerate to regenerate the golden files.
var regenerate = flag.Bool("regenerate", false, "regenerate golden files")

// When the environment variable RUN_AS_PROTOC_GEN_GO is set, we skip running
// tests and instead act as protoc-gen-gogo. This allows the test binary to
// pass itself to protoc.
func init() {
	if os.Getenv("RUN_AS_PROTOC_GEN_GO") != "" {
		main()
		os.Exit(0)
	}
}

func TestGolden(t *testing.T) {
	workdir, err := ioutil.TempDir(protoPath, "proto-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workdir)
	// Find all the proto files we need to compile. We assume that each directory
	// contains the files for a single package.
	packages := map[string][]string{}
	err = filepath.Walk(protoPath, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}
		dir := filepath.Dir(path)
		packages[dir] = append(packages[dir], path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Compile each package, using this binary as protoc-gen-gogo.
	for _, sources := range packages {
		args := []string{
			"--proto_path=.:test/proto",
			fmt.Sprintf("-I=%s/src/", os.Getenv("GOPATH")),
			fmt.Sprintf("-I=%s/src/github.com/protocolbuffers/protobuf/src/", os.Getenv("GOPATH")),
			"--go-sqlmap_out=Msqlgen/sqlgen.proto=github.com/roderm/protoc-gen-go-sqlmap/lib/go/proto/sqlgen/v1:" + workdir}
		args = append(args, sources...)
		protoc(t, args)
	}

	// Compare each generated file to the golden version.
	filepath.Walk(workdir, func(genPath string, info os.FileInfo, _ error) error {
		if info.IsDir() {
			return nil
		}

		// For each generated file, figure out the path to the corresponding
		// golden file in the testdata directory.
		relPath, rerr := filepath.Rel(workdir, genPath)
		t.Logf("...filepath.Rel(%q, %q) = %q \n", workdir, genPath, relPath)
		if rerr != nil {
			t.Errorf("filepath.Rel(%q, %q): %v", workdir, genPath, rerr)
			return nil
		}
		if filepath.SplitList(relPath)[0] == ".." {
			t.Errorf("generated file %q is not relative to %q", genPath, workdir)
		}
		goldenPath := relPath // filepath.Join(protoPath, relPath)

		got, gerr := ioutil.ReadFile(genPath)
		if gerr != nil {
			t.Error(gerr)
			return nil
		}
		if *regenerate {
			// If --regenerate set, just rewrite the golden files.
			err := ioutil.WriteFile(goldenPath, got, 0666)
			if err != nil {
				t.Error(err)
			}
			return nil
		}

		want, err := ioutil.ReadFile(goldenPath)
		if err != nil {
			t.Error(err)
			return nil
		}

		want = fdescRE.ReplaceAll(want, nil)
		got = fdescRE.ReplaceAll(got, nil)
		if bytes.Equal(got, want) {
			return nil
		}

		cmd := exec.Command("diff", "-u", goldenPath, genPath)
		out, _ := cmd.CombinedOutput()
		t.Errorf("golden file differs: %v\n%v", relPath, string(out))
		return nil
	})
}

func protoc(t *testing.T, args []string) {
	cmd := exec.Command(protoExecutable)
	cmd.Args = append(cmd.Args, args...)
	// We set the RUN_AS_PROTOC_GEN_GO environment variable to indicate that
	// the subprocess should act as a proto compiler rather than a test.
	cmd.Env = append(os.Environ(), "RUN_AS_PROTOC_GEN_GO=1")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 || err != nil {
		t.Log("RUNNING: ", strings.Join(cmd.Args, " "))
	}
	if len(out) > 0 {
		t.Log(string(out))
	}
	if err != nil {
		t.Fatalf("protoc: %v", err)
	}
}
