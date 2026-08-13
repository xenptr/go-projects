package validation

import "github.com/xenptr/go-projects/todo-list-api/internal/dto"

func RegisterRequest(r dto.RegisterRequest) error {
	v := New()
	v.Required("name", r.Name)
	v.Required("email", r.Email)
	v.Email("email", r.Email)
	v.Required("password", r.Password)
	v.MinLength("password", r.Password, 8)
	return v.Err()
}

func LoginRequest(r dto.LoginRequest) error {
	v := New()
	v.Required("email", r.Email)
	v.Required("password", r.Password)
	return v.Err()
}
