package handlers

import (
	"errors"
	"fmt"
	"testing"
)

func TestRedirectController(t *testing.T) {
	tests := []struct {
		ext         string
		expectedURL string
		expectedErr error
	}{
		{"", "", ErrExtensionNotFound},
		{"g", "https://www.google.com", nil},
		{"a", "", ErrExtensionNotFound},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("extension-%q", tt.ext), func(t *testing.T) {
			url, err := RedirectController(tt.ext)
			if url != tt.expectedURL {
				t.Errorf("expected URL %q got %q", tt.expectedURL, url)
			}
			if tt.expectedErr == nil && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %q, got %q", tt.expectedErr, err)
			}
		})
	}
}
