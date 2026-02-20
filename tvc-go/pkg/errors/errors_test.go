package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		field   string
		message string
		want    string
	}{
		{"email", "is required", "validation error on field email: is required"},
		{"age", "must be positive", "validation error on field age: must be positive"},
	}

	for _, tt := range tests {
		err := &ValidationError{Field: tt.field, Message: tt.message}
		assert.Equal(t, tt.want, err.Error())
		assert.Implements(t, (*error)(nil), err)
	}
}

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		resource string
		id       string
		want     string
	}{
		{"user", "123", "user not found: 123"},
		{"project", "abc-def", "project not found: abc-def"},
	}

	for _, tt := range tests {
		err := &NotFoundError{Resource: tt.resource, ID: tt.id}
		assert.Equal(t, tt.want, err.Error())
		assert.Implements(t, (*error)(nil), err)
	}
}

func TestConflictError(t *testing.T) {
	tests := []struct {
		resource string
		message  string
		want     string
	}{
		{"project", "slug already exists", "conflict on project: slug already exists"},
		{"user", "email taken", "conflict on user: email taken"},
	}

	for _, tt := range tests {
		err := &ConflictError{Resource: tt.resource, Message: tt.message}
		assert.Equal(t, tt.want, err.Error())
		assert.Implements(t, (*error)(nil), err)
	}
}
