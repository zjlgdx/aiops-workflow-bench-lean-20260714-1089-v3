package main

import (
	"bytes"
	"fmt"
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

func TestCLICompletesAndFiltersTodos(t *testing.T) {
	binary := buildCLI(t)
	database := filepath.Join(t.TempDir(), "todos.json")

	for _, title := range []string{"First", "Second", "Third"} {
		if _, stderr, err := runCLI(binary, database, "add", title); err != nil {
			t.Fatalf("add %q: %v\nstderr: %s", title, err, stderr)
		}
	}

	for attempt := 1; attempt <= 2; attempt++ {
		stdout, stderr, err := runCLI(binary, database, "done", "2")
		if err != nil {
			t.Fatalf("complete todo on attempt %d: %v\nstderr: %s", attempt, err, stderr)
		}
		if stdout != "completed 2\n" {
			t.Fatalf("complete attempt %d stdout = %q, want %q", attempt, stdout, "completed 2\n")
		}
		if stderr != "" {
			t.Fatalf("complete attempt %d stderr = %q, want empty", attempt, stderr)
		}
	}

	stdout, stderr, err := runCLI(binary, database, "list", "--status", "active")
	if err != nil {
		t.Fatalf("list active todos: %v\nstderr: %s", err, stderr)
	}
	if stdout != "1\tactive\tFirst\n3\tactive\tThird\n" {
		t.Fatalf("list active stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("list active stderr = %q, want empty", stderr)
	}

	stdout, stderr, err = runCLI(binary, database, "list", "--status", "done")
	if err != nil {
		t.Fatalf("list done todos: %v\nstderr: %s", err, stderr)
	}
	if stdout != "2\tdone\tSecond\n" {
		t.Fatalf("list done stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("list done stderr = %q, want empty", stderr)
	}
}

func TestCLIRejectsInvalidTodoIDsWithoutModifyingDatabase(t *testing.T) {
	binary := buildCLI(t)

	tests := []struct {
		name       string
		args       []string
		wantStderr string
	}{
		{name: "missing argument", args: []string{"done"}, wantStderr: "todo ID is required\n"},
		{name: "malformed", args: []string{"done", "not-a-number"}, wantStderr: "invalid todo ID \"not-a-number\"\n"},
		{name: "not found", args: []string{"done", "999"}, wantStderr: "todo 999 not found\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database := filepath.Join(t.TempDir(), "todos.json")
			if _, stderr, err := runCLI(binary, database, "add", "Keep active"); err != nil {
				t.Fatalf("seed database: %v\nstderr: %s", err, stderr)
			}
			before, err := os.ReadFile(database)
			if err != nil {
				t.Fatalf("read database before invalid done: %v", err)
			}

			stdout, stderr, err := runCLI(binary, database, tt.args...)
			if err == nil {
				t.Fatal("invalid done command succeeded, want non-zero exit")
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			if stderr != tt.wantStderr {
				t.Fatalf("stderr = %q, want %q", stderr, tt.wantStderr)
			}

			after, err := os.ReadFile(database)
			if err != nil {
				t.Fatalf("read database after invalid done: %v", err)
			}
			if !bytes.Equal(after, before) {
				t.Fatalf("database changed after invalid done:\nbefore: %s\nafter:  %s", before, after)
			}
		})
	}
}

func TestCLIRejectsUnsupportedStatusWithoutModifyingDatabase(t *testing.T) {
	binary := buildCLI(t)

	for _, status := range []string{"blocked", ""} {
		t.Run(status, func(t *testing.T) {
			database := filepath.Join(t.TempDir(), "todos.json")
			if _, stderr, err := runCLI(binary, database, "add", "Keep active"); err != nil {
				t.Fatalf("seed database: %v\nstderr: %s", err, stderr)
			}
			before, err := os.ReadFile(database)
			if err != nil {
				t.Fatalf("read database before invalid list: %v", err)
			}

			stdout, stderr, err := runCLI(binary, database, "list", "--status", status)
			if err == nil {
				t.Fatal("unsupported status succeeded, want non-zero exit")
			}
			if stdout != "" {
				t.Fatalf("unsupported status stdout = %q, want empty", stdout)
			}
			wantStderr := fmt.Sprintf("unsupported status %q\n", status)
			if stderr != wantStderr {
				t.Fatalf("unsupported status stderr = %q, want %q", stderr, wantStderr)
			}

			after, err := os.ReadFile(database)
			if err != nil {
				t.Fatalf("read database after invalid list: %v", err)
			}
			if !bytes.Equal(after, before) {
				t.Fatalf("database changed after invalid list:\nbefore: %s\nafter:  %s", before, after)
			}
		})
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
