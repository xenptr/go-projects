package main

import "fmt"

func Help() {
	fmt.Print(`
Task Tracker CLI

Usage:
  ./task-cli <command> [arguments]

Commands:
  add <description>
      Add a new task

  update <id> <description>
      Update a task description

  delete <id>
      Delete a task

  mark-in-progress <id>
      Mark a task as in progress

  mark-done <id>
      Mark a task as done

  list
      List all tasks

  list <status>
      List tasks by status

Statuses:
  todo
  in-progress
  done

Examples:
  ./task-cli add "Learn Go"
  ./task-cli update 1 "Learn Go deeply"
  ./task-cli delete 1
  ./task-cli mark-in-progress 1
  ./task-cli mark-done 1
  ./task-cli list
  ./task-cli list done
`)
}
