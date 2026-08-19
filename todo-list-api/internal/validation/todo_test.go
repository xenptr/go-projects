package validation

import (
	"testing"

	"github.com/xenptr/go-projects/todo-list-api/internal/dto"
)

func TestCreateTodoRequest(t *testing.T) {
	tests := []struct {
		name    string
		request dto.CreateTodoRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: dto.CreateTodoRequest{
				Title: "Buy groceries",
			},
			wantErr: false,
		},
		{
			name:    "missing title",
			request: dto.CreateTodoRequest{},
			wantErr: true,
		},
		{
			name: "whitespace-only title",
			request: dto.CreateTodoRequest{
				Title: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateTodoRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateTodoRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateTodoRequest(t *testing.T) {
	tests := []struct {
		name    string
		request dto.UpdateTodoRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: dto.UpdateTodoRequest{
				Title: "Buy groceries",
			},
			wantErr: false,
		},
		{
			name:    "missing title",
			request: dto.UpdateTodoRequest{},
			wantErr: true,
		},
		{
			name: "whitespace-only title",
			request: dto.UpdateTodoRequest{
				Title: "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UpdateTodoRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateTodoRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
