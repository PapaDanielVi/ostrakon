package vault

import (
	"testing"
)

func TestFileInfo(t *testing.T) {
	file := FileInfo{
		Path: "secrets/api-key",
		SHA:  "abc123def456",
		Size: 100,
	}

	if file.Path != "secrets/api-key" {
		t.Errorf("Path = %v, want secrets/api-key", file.Path)
	}
	if file.SHA != "abc123def456" {
		t.Errorf("SHA = %v, want abc123def456", file.SHA)
	}
	if file.Size != 100 {
		t.Errorf("Size = %v, want 100", file.Size)
	}
}

// TestPathFiltering tests the path stripping logic used in ListFiles
func TestPathFiltering(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "regular path",
			path:     "contents/secret.txt",
			expected: "secret.txt",
		},
		{
			name:     "nested path",
			path:     "contents/profiles/prod/db",
			expected: "profiles/prod/db",
		},
		{
			name:     "short path - no prefix",
			path:     "x",
			expected: "x",
		},
		{
			name:     "exact prefix length",
			path:     "contents/",
			expected: "contents/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripContentsPrefix(tt.path)

			if result != tt.expected {
				t.Errorf("stripContentsPrefix(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

// stripContentsPrefix removes the "contents/" prefix if present
// This mirrors the logic in ListFiles
func stripContentsPrefix(path string) string {
	if len(path) > 9 && path[:9] == "contents/" {
		return path[9:]
	}
	return path
}

// TestProfilePath tests building profile paths
func TestProfilePath(t *testing.T) {
	tests := []struct {
		name      string
		profile   string
		secret    string
		expected  string
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
		{
			name:     "nested with profile",
			profile:  "dev",
			secret:   "secrets/api",
			expected: "profiles/dev/secrets/api",
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

// buildVaultPath constructs the vault path for a secret
func buildVaultPath(profile, secret string) string {
	if profile == "" {
		return secret
	}
	return "profiles/" + profile + "/" + secret
}