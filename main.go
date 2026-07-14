package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type todo struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Title  string `json:"title"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "version" {
		fmt.Fprintln(stdout, "todo-bench seed")
		return 0
	}

	database := os.Getenv("TODO_DB")
	switch {
	case len(args) == 2 && args[0] == "add":
		return addTodo(database, strings.TrimSpace(args[1]), stdout, stderr)
	case len(args) == 1 && args[0] == "list":
		return listTodos(database, "", false, stdout, stderr)
	case len(args) == 3 && args[0] == "list" && args[1] == "--status":
		return listTodos(database, args[2], true, stdout, stderr)
	case len(args) == 1 && args[0] == "done":
		fmt.Fprintln(stderr, "todo ID is required")
		return 2
	case len(args) == 2 && args[0] == "done":
		id, err := strconv.Atoi(args[1])
		if err != nil || id < 1 {
			fmt.Fprintf(stderr, "invalid todo ID %q\n", args[1])
			return 2
		}
		return completeTodo(database, id, stdout, stderr)
	default:
		fmt.Fprintln(stderr, "usage: todo <add|list|done>")
		return 2
	}
}

func addTodo(database, title string, stdout, stderr io.Writer) int {
	if title == "" {
		fmt.Fprintln(stderr, "title must not be empty")
		return 1
	}

	todos, err := readTodos(database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	nextID := 1
	for _, item := range todos {
		if item.ID >= nextID {
			nextID = item.ID + 1
		}
	}
	todos = append(todos, todo{ID: nextID, Status: "active", Title: title})

	if err := writeTodos(database, todos); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "added %d\n", nextID)
	return 0
}

func completeTodo(database string, id int, stdout, stderr io.Writer) int {
	todos, err := readTodos(database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	for i := range todos {
		if todos[i].ID != id {
			continue
		}
		if todos[i].Status != "done" {
			todos[i].Status = "done"
			if err := writeTodos(database, todos); err != nil {
				fmt.Fprintln(stderr, err)
				return 1
			}
		}
		fmt.Fprintf(stdout, "completed %d\n", id)
		return 0
	}

	fmt.Fprintf(stderr, "todo %d not found\n", id)
	return 1
}

func listTodos(database, status string, filter bool, stdout, stderr io.Writer) int {
	if filter && status != "active" && status != "done" {
		fmt.Fprintf(stderr, "unsupported status %q\n", status)
		return 2
	}

	todos, err := readTodos(database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, item := range todos {
		if filter && item.Status != status {
			continue
		}
		fmt.Fprintf(stdout, "%d\t%s\t%s\n", item.ID, item.Status, item.Title)
	}
	return 0
}

func readTodos(database string) ([]todo, error) {
	data, err := os.ReadFile(database)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var todos []todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func writeTodos(database string, todos []todo) error {
	data, err := json.Marshal(todos)
	if err != nil {
		return err
	}
	return os.WriteFile(database, data, 0o600)
}
