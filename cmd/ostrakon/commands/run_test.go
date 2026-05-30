package commands

import (
	"errors"
	"strings"
	"testing"
)

func TestParseSecretEnvFormat(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		expectKey string
		expectVal string
		expectErr bool
	}{
		{
			name:      "valid key value",
			secret:    "API_KEY=secret123",
			expectKey: "API_KEY",
			expectVal: "secret123",
			expectErr: false,
		},
		{
			name:      "value with equals",
			secret:    "DATABASE_URL=postgres://user:pass@host/db?ssl=true",
			expectKey: "DATABASE_URL",
			expectVal: "postgres://user:pass@host/db?ssl=true",
			expectErr: false,
		},
		{
			name:     "missing equals",
			secret:   "invalid_format",
			expectErr: true,
		},
		{
			name:     "empty secret",
			secret:   "",
			expectErr: true,
		},
		{
			name:      "key only no value",
			secret:    "API_KEY=",
			expectKey: "API_KEY",
			expectVal: "",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, val, err := parseSecretEnv(tt.secret)

			if tt.expectErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if key != tt.expectKey {
				t.Errorf("key = %q, want %q", key, tt.expectKey)
			}
			if val != tt.expectVal {
				t.Errorf("val = %q, want %q", val, tt.expectVal)
			}
		})
	}
}

func parseSecretEnv(secret string) (key, value string, err error) {
	if secret == "" {
		return "", "", errors.New("secret cannot be empty")
	}
	parts := strings.SplitN(secret, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.New("secret is not in KEY=VALUE format")
	}
	return parts[0], parts[1], nil
}