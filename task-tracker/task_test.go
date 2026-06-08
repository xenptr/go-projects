package main

import "testing"

func resetTasks() {
	tasks = nil
}

func createTask(t *testing.T, description string) int {
	t.Helper()

	id := AddTask(description)

	if len(tasks) == 0 {
		t.Fatal("task was not added")
	}

	return id
}

func TestAddTask(t *testing.T) {
	resetTasks()

	id := AddTask("Learn Go")

	if id != 1 {
		t.Errorf("expected ID 1, got %d", id)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	task := tasks[0]

	if task.Description != "Learn Go" {
		t.Errorf("expected description %q, got %q", "Learn Go", task.Description)
	}

	if task.Status != Todo {
		t.Errorf("expected status %q, got %q", Todo, task.Status)
	}
}

func TestUpdateTask(t *testing.T) {
	resetTasks()

	id := createTask(t, "Old")

	err := UpdateTask(id, "New")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tasks[0].Description != "New" {
		t.Errorf("expected New, got %q", tasks[0].Description)
	}
}

func TestDeleteTask(t *testing.T) {
	resetTasks()

	id := createTask(t, "Delete Me")

	if err := DeleteTask(id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestSetTaskStatus(t *testing.T) {
	resetTasks()

	id := createTask(t, "Learn Testing")

	if err := SetTaskStatus(id, Done); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tasks[0].Status != Done {
		t.Errorf("expected status %q, got %q", Done, tasks[0].Status)
	}
}

func TestListTasks(t *testing.T) {
	resetTasks()

	AddTask("Task 1")
	AddTask("Task 2")

	got := ListTasks()

	if len(got) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(got))
	}

	if got[0].Description != "Task 1" {
		t.Errorf("expected Task 1, got %s", got[0].Description)
	}

	if got[1].Description != "Task 2" {
		t.Errorf("expected Task 2, got %s", got[1].Description)
	}
}

func TestListTasksByStatus(t *testing.T) {
	resetTasks()

	id1 := createTask(t, "Task 1")
	id2 := createTask(t, "Task 2")

	_ = SetTaskStatus(id1, Done)
	_ = SetTaskStatus(id2, InProgress)

	got, err := ListTasksByStatus(Done)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got))
	}

	if got[0].Status != Done {
		t.Errorf("expected status %q, got %q", Done, got[0].Status)
	}
}

func TestListTasksByStatusInvalid(t *testing.T) {
	resetTasks()

	if _, err := ListTasksByStatus(Status("banana")); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Status
		wantErr bool
	}{
		{"todo", "todo", Todo, false},
		{"done", "done", Done, false},
		{"in-progress", "in-progress", InProgress, false},
		{"invalid", "pending", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseStatus(tc.input)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestFindTaskIndex(t *testing.T) {
	resetTasks()

	id := createTask(t, "Find Me")

	index, err := findTaskIndex(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if index != 0 {
		t.Errorf("expected index 0, got %d", index)
	}
}

func TestFindTaskIndexNotFound(t *testing.T) {
	resetTasks()

	if _, err := findTaskIndex(999); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNotFoundErrors(t *testing.T) {
	resetTasks()

	tests := []struct {
		name string
		fn   func() error
	}{
		{"UpdateTask", func() error {
			return UpdateTask(999, "test")
		}},
		{"DeleteTask", func() error {
			return DeleteTask(999)
		}},
		{"SetTaskStatus", func() error {
			return SetTaskStatus(999, Done)
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
