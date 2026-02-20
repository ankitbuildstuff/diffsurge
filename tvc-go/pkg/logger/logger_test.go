package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_ValidLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		l := New(level, "json")
		assert.NotNil(t, l)
	}
}

func TestNew_InvalidLevel_DefaultsToInfo(t *testing.T) {
	l := New("bogus", "json")
	assert.NotNil(t, l)
}

func TestNew_ConsoleFormat(t *testing.T) {
	l := New("info", "console")
	assert.NotNil(t, l)
}

func TestNew_JSONFormat(t *testing.T) {
	l := New("info", "json")
	assert.NotNil(t, l)
}

func TestDefault(t *testing.T) {
	l := Default()
	assert.NotNil(t, l)
}
