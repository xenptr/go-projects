package validation

import "github.com/xenptr/go-projects/todo-list-api/internal/dto"

func CreateTodoRequest(r dto.CreateTodoRequest) error {
	v := New()
	v.Required("title", r.Title)
	return v.Err()
}

func UpdateTodoRequest(r dto.UpdateTodoRequest) error {
	v := New()
	v.Required("title", r.Title)
	return v.Err()
}
