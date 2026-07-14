package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
		return listTodos(database, stdout, stderr)
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

	data, err := json.Marshal(todos)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := os.WriteFile(database, data, 0o600); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "added %d\n", nextID)
	return 0
}

func listTodos(database string, stdout, stderr io.Writer) int {
	todos, err := readTodos(database)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	for _, item := range todos {
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
