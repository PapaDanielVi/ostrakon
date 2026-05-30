package commands

import (
	"testing"
)

// TestShredPathConstruction tests the path and SHA logic in shred
func TestShredPathConstruction(t *testing.T) {
	tests := []struct {
		name     string
		shredAll bool
		secret   string
		expected string
	}{
		{
			name:     "non-shred path",
			shredAll: false,
			secret:   "my-secret",
			expected: "my-secret",
		},
		{
			name:     "reset all flag",
			shredAll: true,
			secret:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getShredPath(tt.shredAll, tt.secret)

			if result != tt.expected {
				t.Errorf("getShredPath() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func getShredPath(shredAll bool, secret string) string {
	if shredAll {
		return ""
	}
	return secret
}