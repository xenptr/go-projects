package validation

import (
	"strings"
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{
		Errors: map[string]string{
			"email": "must be a valid email address",
		},
	}

	errMsg := ve.Error()
	if !strings.Contains(errMsg, "email: must be a valid email address") {
		t.Errorf("unexpected error message: %s", errMsg)
	}
}

func TestValidator_Required(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{
			name:      "empty string fails",
			field:     "name",
			value:     "",
			wantError: true,
		},
		{
			name:      "whitespace only fails",
			field:     "name",
			value:     "   ",
			wantError: true,
		},
		{
			name:      "non empty string passes",
			field:     "name",
			value:     "John Doe",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Required(tt.field, tt.value)

			if tt.wantError {
				if v.OK() {
					t.Error("expected validator.OK() to be false")
				}
				if v.Err() == nil {
					t.Error("expected validator.Err() to return an error")
				}
			} else {
				if !v.OK() {
					t.Errorf("expected validator.OK() to be true, got errors: %v", v.errors)
				}
				if v.Err() != nil {
					t.Errorf("expected validator.Err() to be nil, got: %v", v.Err())
				}
			}
		})
	}
}

func TestValidator_MinLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		min       int
		wantError bool
	}{
		{
			name:      "shorter than min fails",
			field:     "password",
			value:     "12345",
			min:       6,
			wantError: true,
		},
		{
			name:      "exact length passes",
			field:     "password",
			value:     "123456",
			min:       6,
			wantError: false,
		},
		{
			name:      "longer than min passes",
			field:     "password",
			value:     "12345678",
			min:       6,
			wantError: false,
		},
		{
			name:      "whitespace padding not counted",
			field:     "password",
			value:     "   12345   ",
			min:       6,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.MinLength(tt.field, tt.value, tt.min)

			if tt.wantError && v.OK() {
				t.Error("expected validation failure")
			}
			if !tt.wantError && !v.OK() {
				t.Errorf("expected validation success, got: %v", v.errors)
			}
		})
	}
}

func TestValidator_Email(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{
			name:      "empty email skipped",
			field:     "email",
			value:     "",
			wantError: false,
		},
		{
			name:      "whitespace email skipped",
			field:     "email",
			value:     "   ",
			wantError: false,
		},
		{
			name:      "invalid email fails",
			field:     "email",
			value:     "invalid-email",
			wantError: true,
		},
		{
			name:      "valid email passes",
			field:     "email",
			value:     "test@example.com",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Email(tt.field, tt.value)

			if tt.wantError && v.OK() {
				t.Error("expected validation failure")
			}
			if !tt.wantError && !v.OK() {
				t.Errorf("expected validation success, got: %v", v.errors)
			}
		})
	}
}
