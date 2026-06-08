package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	var command string

	if err := LoadTasks(); err != nil {
		fmt.Println(err)
		return
	}

	if len(os.Args) < 2 {
		Help()
		return
	}
	command = os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 3 {
			fmt.Println("Usage: add <description>")
			return
		}

		description := strings.Join(os.Args[2:], " ")
		id := AddTask(description)
		if err := SaveTasks(); err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Task added successfully (ID: %v)\n", id)
		return
	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: update <id> <description>")
			return
		}

		id, err := parseID(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		description := strings.Join(os.Args[3:], " ")
		if err := UpdateTask(id, description); err != nil {
			fmt.Println(err)
			return
		}
		if err := SaveTasks(); err != nil {
			fmt.Println(err)
			return
		}
		return
	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("Usage: delete <id>")
			return
		}

		id, err := parseID(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := DeleteTask(id); err != nil {
			fmt.Println(err)
			return
		}
		if err := SaveTasks(); err != nil {
			fmt.Println(err)
			return
		}
		return
	case "mark-in-progress":
		if len(os.Args) != 3 {
			fmt.Println("Usage: mark-in-progress <id>")
			return
		}

		id, err := parseID(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := SetTaskStatus(id, InProgress); err != nil {
			fmt.Println(err)
			return
		}
		if err := SaveTasks(); err != nil {
			fmt.Println(err)
			return
		}
		return
	case "mark-done":
		if len(os.Args) != 3 {
			fmt.Println("Usage: mark-done <id>")
			return
		}

		id, err := parseID(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}

		if err := SetTaskStatus(id, Done); err != nil {
			fmt.Println(err)
			return
		}
		if err := SaveTasks(); err != nil {
			fmt.Println(err)
			return
		}
		return
	case "list":
		if len(os.Args) == 2 {
			PrintTaskList(ListTasks())
			return
		}

		if len(os.Args) == 3 {
			status, err := ParseStatus(os.Args[2])
			if err != nil {
				fmt.Println(err)
				return
			}

			filtered, err := ListTasksByStatus(status)
			if err != nil {
				fmt.Println(err)
				return
			}
			PrintTaskList(filtered)
			return
		}

		fmt.Println("Usage: list [status]")
		return
	case "help":
		Help()
		return
	default:
		fmt.Print("Unknown Command")
		return
	}
}

func parseID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: must be a number")
	}

	return id, nil
}
