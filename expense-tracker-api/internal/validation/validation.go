package validation

import (
	"fmt"
	"net/mail"
	"strings"
)

// ValidationError holds a map of field -> error message.
type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	msgs := make([]string, 0, len(e.Errors))
	for field, msg := range e.Errors {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(msgs, "; ")
}

// validator accumulates field errors and reports if any exist.
type validator struct {
	errors map[string]string
}

func New() *validator {
	return &validator{errors: make(map[string]string)}
}

func (v *validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors[field] = "is required"
	}
}

func (v *validator) MinLength(field, value string, min int) {
	if len(strings.TrimSpace(value)) < min {
		v.errors[field] = fmt.Sprintf("must be at least %d characters", min)
	}
}

func (v *validator) Email(field, value string) {
	if strings.TrimSpace(value) == "" {
		return
	}

	if _, err := mail.ParseAddress(value); err != nil {
		v.errors[field] = "must be a valid email address"
	}
}

func (v *validator) OK() bool {
	return len(v.errors) == 0
}

func (v *validator) Err() error {
	if v.OK() {
		return nil
	}
	return &ValidationError{Errors: v.errors}
}
