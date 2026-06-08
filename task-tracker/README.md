# Task Tracker CLI

A simple command-line task tracker built with Go. This application allows you to manage tasks directly from the terminal by adding, updating, deleting, and tracking their status.

## Project URL

https://roadmap.sh/projects/task-tracker

## Features

* Add new tasks
* Update existing tasks
* Delete tasks
* Mark tasks as:

  * `todo`
  * `in-progress`
  * `done`
* List all tasks
* Filter tasks by status
* Persistent storage using JSON

## Installation

Clone the repository:

```bash
git clone <repository-url>
cd task-tracker
```

Run the application:

```bash
go run .
```

Or build an executable:

```bash
go build -o task-cli
```

Run the executable:

```bash
./task-cli
```

## Usage

### Add a task

```bash
./task-cli add "Learn Go"
```

### Update a task

```bash
./task-cli update 1 "Learn Go Testing"
```

### Delete a task

```bash
./task-cli delete 1
```

### Mark a task as in progress

```bash
./task-cli mark-in-progress 1
```

### Mark a task as done

```bash
./task-cli mark-done 1
```

### List all tasks

```bash
./task-cli list
```

### Display available commands

```bash
./task-cli help
```


### List tasks by status

```bash
./task-cli list done
./task-cli list todo
./task-cli list in-progress
```

## Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

## Project Structure

```text
.
├── main.go         # CLI entry point and command routing
├── task.go         # Task operations and business logic
├── storage.go      # JSON file persistence
├── help.go         # Help menu and usage information
├── task_test.go    # Unit tests
├── tasks.json      # Task storage file
└── README.md
```

