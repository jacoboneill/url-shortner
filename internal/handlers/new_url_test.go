package handlers

import (
	"errors"
	"testing"
)

func TestNewURLController(t *testing.T) {
	tests := []struct {
		url           string
		expectedError error
	}{
		{"https://api.kanye.rest", nil},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if _, err := NewURLController(tt.url); !errors.Is(err, tt.expectedError) {
				t.Errorf("expected %v, got %v", tt.expectedError, err)
			}
		})
	}
}
