package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLIAddAndListTodos(t *testing.T) {
	binary := buildCLI(t)
	database := filepath.Join(t.TempDir(), "todos.json")

	stdout, stderr, err := runCLI(binary, database, "add", "  Buy milk  ")
	if err != nil {
		t.Fatalf("add first todo: %v\nstderr: %s", err, stderr)
	}
	if stdout != "added 1\n" {
		t.Fatalf("add first todo stdout = %q, want %q", stdout, "added 1\n")
	}
	if stderr != "" {
		t.Fatalf("add first todo stderr = %q, want empty", stderr)
	}

	stdout, stderr, err = runCLI(binary, database, "add", "Call mom")
	if err != nil {
		t.Fatalf("add second todo: %v\nstderr: %s", err, stderr)
	}
	if stdout != "added 2\n" {
		t.Fatalf("add second todo stdout = %q, want %q", stdout, "added 2\n")
	}
	if stderr != "" {
		t.Fatalf("add second todo stderr = %q, want empty", stderr)
	}

	stdout, stderr, err = runCLI(binary, database, "list")
	if err != nil {
		t.Fatalf("list todos: %v\nstderr: %s", err, stderr)
	}
	if stdout != "1\tactive\tBuy milk\n2\tactive\tCall mom\n" {
		t.Fatalf("list stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("list stderr = %q, want empty", stderr)
	}
}

func TestCLIRejectsEmptyTitleWithoutModifyingDatabase(t *testing.T) {
	binary := buildCLI(t)
	database := filepath.Join(t.TempDir(), "todos.json")

	_, stderr, err := runCLI(binary, database, "add", "Keep me")
	if err != nil {
		t.Fatalf("seed database: %v\nstderr: %s", err, stderr)
	}
	before, err := os.ReadFile(database)
	if err != nil {
		t.Fatalf("read database before invalid add: %v", err)
	}

	stdout, stderr, err := runCLI(binary, database, "add", " \t\n ")
	if err == nil {
		t.Fatal("empty title succeeded, want non-zero exit")
	}
	if stdout != "" {
		t.Fatalf("empty title stdout = %q, want empty", stdout)
	}
	if stderr != "title must not be empty\n" {
		t.Fatalf("empty title stderr = %q, want %q", stderr, "title must not be empty\n")
	}

	after, err := os.ReadFile(database)
	if err != nil {
		t.Fatalf("read database after invalid add: %v", err)
	}
	if !bytes.Equal(after, before) {
		t.Fatalf("database changed after invalid add:\nbefore: %s\nafter:  %s", before, after)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "todo")
	cmd := exec.Command("go", "build", "-o", binary, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	return binary
}

func runCLI(binary, database string, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "TODO_DB="+database)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
