package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAll(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "a")
}

func TestIsLowerCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"строчная буква", "starting server", true},
		{"заглавная буква", "Starting server", false},
		{"пустая строка", "", true},
		{"цифра в начале (допустимо)", "123 error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLowerCase(tt.input)
			assert.Equal(t, tt.expected, got, "input = %q", tt.input)
		})
	}
}

func TestIsEnglish(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"только английский", "database connected", true},
		{"английский с цифрами", "port 8080", true},
		{"русский текст", "ошибка", false},
		{"смешанный текст", "error ошибка", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEnglish(tt.input)
			assert.Equal(t, tt.expected, got, "input = %q", tt.input)
		})
	}
}

func TestHasSpecialCharsOrEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"обычный текст", "server started", false},
		{"с восклицательным знаком", "failed!!!", true},
		{"с точкой", "wait...", true},
		{"с эмодзи", "fire 🔥", true},
		{"математический символ", "a + b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSpecialCharsOrEmoji(tt.input)
			assert.Equal(t, tt.expected, got, "input = %q", tt.input)
		})
	}
}
