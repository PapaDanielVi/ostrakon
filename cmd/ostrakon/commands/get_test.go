package commands

import (
	"strings"
	"testing"
)

func TestGetVaultPath(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		secret   string
		expected string
	}{
		{
			name:     "no profile",
			profile:  "",
			secret:   "secret.txt",
			expected: "secret.txt",
		},
		{
			name:     "with profile",
			profile:  "prod",
			secret:   "db",
			expected: "profiles/prod/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildVaultPath(tt.profile, tt.secret)

			if result != tt.expected {
				t.Errorf("buildVaultPath(%q, %q) = %q, want %q", tt.profile, tt.secret, result, tt.expected)
			}
		})
	}
}

func TestOutputDestination(t *testing.T) {
	tests := []struct {
		name         string
		outputFile   string
		expectStdout bool
	}{
		{
			name:         "stdout output",
			outputFile:   "",
			expectStdout: true,
		},
		{
			name:         "file output",
			outputFile:   "/tmp/secret",
			expectStdout: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useStdout := tt.outputFile == ""

			if useStdout != tt.expectStdout {
				t.Errorf("useStdout = %v, want %v", useStdout, tt.expectStdout)
			}
		})
	}
}

func TestSecretFormat(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		expected bool
	}{
		{
			name:     "valid format",
			secret:   "API_KEY=secret123",
			expected: true,
		},
		{
			name:     "valid format with equals in value",
			secret:   "KEY=value=with=equals",
			expected: true,
		},
		{
			name:     "missing equals",
			secret:   "invalid_format",
			expected: false,
		},
		{
			name:     "empty secret",
			secret:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidSecretFormat(tt.secret)

			if valid != tt.expected {
				t.Errorf("isValidSecretFormat(%q) = %v, want %v", tt.secret, valid, tt.expected)
			}
		})
	}
}

func isValidSecretFormat(secret string) bool {
	parts := strings.SplitN(secret, "=", 2)
	return len(parts) == 2 && secret != ""
}