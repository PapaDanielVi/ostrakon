package commands

import (
	"testing"
)

// getSecretName determines the secret name from args or flags
func getSecretName(args []string, flagName string) string {
	name := flagName
	if name == "" {
		if len(args) > 0 {
			name = args[0]
		}
	}
	return name
}

// buildVaultPath constructs the vault path for a secret - extracted for testing
func buildVaultPath(profile, name string) string {
	if profile == "" {
		return name
	}
	return "profiles/" + profile + "/" + name
}

func TestVaultPathWithProfile(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		secret  string
		expected string
	}{
		{
			name:    "no profile",
			profile: "",
			secret:  "secret.txt",
			expected: "secret.txt",
		},
		{
			name:    "with profile",
			profile: "prod",
			secret:  "db",
			expected: "profiles/prod/db",
		},
		{
			name:    "nested secret with profile",
			profile: "dev",
			secret:  "secrets/api-key",
			expected: "profiles/dev/secrets/api-key",
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

func TestSecretNameFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		expected string
	}{
		{
			name:     "flag takes precedence",
			args:     []string{"file.txt"},
			flagName: "custom-name",
			expected: "custom-name",
		},
		{
			name:     "use arg when no flag",
			args:     []string{"secret.txt"},
			flagName: "",
			expected: "secret.txt",
		},
		{
			name:     "no input",
			args:     []string{},
			flagName: "",
			expected: "",
		},
		{
			name:     "empty flag uses arg",
			args:     []string{"secret.txt"},
			flagName: "",
			expected: "secret.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSecretName(tt.args, tt.flagName)

			if result != tt.expected {
				t.Errorf("getSecretName() = %q, want %q", result, tt.expected)
			}
		})
	}
}