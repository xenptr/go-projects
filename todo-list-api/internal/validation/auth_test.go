package validation

import (
	"testing"

	"github.com/xenptr/go-projects/todo-list-api/internal/dto"
)

func TestRegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		request dto.RegisterRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			request: dto.RegisterRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: true,
		}, {
			name: "whitespace-only name",
			request: dto.RegisterRequest{
				Name:     "   ",
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing email",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "short password",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "short",
			},
			wantErr: true,
		}, {
			name: "missing password",
			request: dto.RegisterRequest{
				Name:  "John Doe",
				Email: "john@example.com",
			},
			wantErr: true,
		}, {
			name: "password exactly minimum length",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "12345678",
			},
			wantErr: false,
		}, {
			name: "password one character below minimum",
			request: dto.RegisterRequest{
				Name:     "John Doe",
				Email:    "john@example.com",
				Password: "1234567",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RegisterRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RegisterRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		request dto.LoginRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: dto.LoginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			request: dto.LoginRequest{
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			request: dto.LoginRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			wantErr: true,
		}, {
			name: "missing password",
			request: dto.LoginRequest{
				Email: "john@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := LoginRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoginRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRefreshRequest(t *testing.T) {
	tests := []struct {
		name    string
		request dto.RefreshRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			request: dto.RefreshRequest{RefreshToken: "some.jwt.token"},
			wantErr: false,
		},
		{
			name:    "missing refresh_token",
			request: dto.RefreshRequest{},
			wantErr: true,
		},
		{
			name:    "whitespace-only refresh_token",
			request: dto.RefreshRequest{RefreshToken: "   "},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RefreshRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RefreshRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
