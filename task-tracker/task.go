package main

import (
	"fmt"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Status      Status `json:"status,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type Status string

const (
	Todo       Status = "todo"
	Done       Status = "done"
	InProgress Status = "in-progress"
)

var tasks []Task

func AddTask(desc string) int {
	id := getNextID()
	now := currentTime()
	task := Task{
		ID:          id,
		Description: desc,
		Status:      Todo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	tasks = append(tasks, task)

	return task.ID
}

func UpdateTask(id int, desc string) error {
	index, err := findTaskIndex(id)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	tasks[index].Description = desc
	tasks[index].UpdatedAt = currentTime()

	return nil
}

func DeleteTask(id int) error {
	index, err := findTaskIndex(id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	tasks = append(tasks[:index], tasks[index+1:]...)

	return nil
}

func SetTaskStatus(id int, status Status) error {
	index, err := findTaskIndex(id)
	if err != nil {
		return fmt.Errorf("set status: %w", err)
	}

	tasks[index].Status = status
	tasks[index].UpdatedAt = currentTime()

	return nil
}

func findTaskIndex(id int) (int, error) {
	for i := range tasks {
		if tasks[i].ID == id {
			return i, nil
		}
	}

	return -1, fmt.Errorf("task %d not found", id)
}

func getNextID() int {
	maxID := 0

	for i := range tasks {
		if tasks[i].ID > maxID {
			maxID = tasks[i].ID
		}
	}

	return maxID + 1
}

func currentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func ListTasks() []Task {
	return tasks
}

func ListTasksByStatus(status Status) ([]Task, error) {
	if !isValidStatus(status) {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	filtered := []Task{}

	for _, task := range tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}

	return filtered, nil
}

func PrintTaskList(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return
	}

	fmt.Printf("%-4s %-12s %-30s\n", "ID", "STATUS", "DESCRIPTION")

	for _, task := range tasks {
		fmt.Printf("%-4d %-12s %-30s\n", task.ID, task.Status, task.Description)
	}
}

func isValidStatus(status Status) bool {
	switch status {
	case Todo, InProgress, Done:
		return true
	default:
		return false
	}
}

func ParseStatus(s string) (Status, error) {
	switch s {
	case string(Todo):
		return Todo, nil
	case string(Done):
		return Done, nil
	case string(InProgress):
		return InProgress, nil
	default:
		return "", fmt.Errorf("invalid status: %s", s)
	}
}

