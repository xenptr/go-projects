package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	Done = "done"

	InProgress = "in-progress"
)

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

var tasks []Task

func CreateTask(desc string) int {
	id := nextID()
	task := Task{
		ID:          id,
		Description: desc,
		Status:      "todo",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	tasks = append(tasks, task)

	return task.ID
}

func Read() {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(data))
}

func Update(id int, desc string) {
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Description = desc
			tasks[i].UpdatedAt = time.Now()
			return
		}
	}
}

func Delete(id int) {
	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return
		}
	}
}

func Mark(id int, status string) {
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Status = status
			return
		}
	}
}

func nextID() int {
	maxID := 0

	for i := range tasks {
		if tasks[i].ID > maxID {
			maxID = tasks[i].ID
		}
	}

	return maxID + 1
}

func LoadTasks() {
	file, err := os.ReadFile("tasks.json")
	if err != nil {
		fmt.Println(err)
	}

	if errors.Is(err, os.ErrNotExist) {
		tasks = []Task{}
		return
	}

	err = json.Unmarshal(file, &tasks)
	if err != nil {
		fmt.Println(err)
	}
}

func SaveTasks() {
	file, err := os.OpenFile("tasks.json", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		fmt.Println(err)
	}

	_, err = file.Write(data)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	var command string

	LoadTasks()
	defer SaveTasks()

	if len(os.Args) < 2 {
		fmt.Println("Help menu")
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
		id := CreateTask(description)
		fmt.Printf("Task added successfully (ID: %v)\n", id)
		return
	case "update":
		if len(os.Args) < 4 {
			fmt.Println("Usage: update <id> <description>")
			return
		}

		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Print("ID must be an number")
			return
		}
		description := strings.Join(os.Args[3:], " ")

		Update(id, description)
	case "delete":
		if len(os.Args) != 3 {
			fmt.Println("Usage: delete <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Print("ID must be an number")
			return
		}
		Delete(id)
	case "mark-in-progress":
		if len(os.Args) != 3 {
			fmt.Println("Usage: mark-in-progress <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Print("ID must be an number")
			return
		}
		Mark(id, InProgress)
		return
	case "mark-done":
		if len(os.Args) != 3 {
			fmt.Println("Usage: mark-done <id>")
			return
		}
		id, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Print("ID must be an number")
			return
		}
		Mark(id, Done)
		return
	case "list":
		if len(os.Args) == 2 {
			Read()
			return
		} else if len(os.Args) == 3 {
			status := os.Args[2]
			for i := range tasks {
				if tasks[i].Status == status {
					fmt.Println(tasks[i])
				}
			}
			return
		} else {
			fmt.Println("Usage: list [status]")
		}
	case "help":
		fmt.Println("help menu")
	default:
		fmt.Print("Unknown Command")
	}
}
