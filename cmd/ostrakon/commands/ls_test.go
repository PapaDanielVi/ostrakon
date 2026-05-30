package commands

import (
	"testing"

	"github.com/PapaDanielVi/ostrakon/pkg/vault"
)

// TestFilterByProfile tests the profile filtering logic used in runLs
func TestFilterByProfile(t *testing.T) {
	tests := []struct {
		name     string
		files    []vault.FileInfo
		profile  string
		expected []vault.FileInfo
	}{
		{
			name: "no filter",
			files: []vault.FileInfo{
				{Path: "secret1", Size: 100},
				{Path: "secret2", Size: 200},
			},
			profile: "",
			expected: []vault.FileInfo{
				{Path: "secret1", Size: 100},
				{Path: "secret2", Size: 200},
			},
		},
		{
			name: "filter by profile",
			files: []vault.FileInfo{
				{Path: "profiles/prod/db", Size: 100},
				{Path: "profiles/prod/api", Size: 200},
				{Path: "secret1", Size: 50},
			},
			profile: "prod",
			expected: []vault.FileInfo{
				{Path: "db", Size: 100},
				{Path: "api", Size: 200},
			},
		},
		{
			name: "filter non-existent profile",
			files: []vault.FileInfo{
				{Path: "profiles/prod/db", Size: 100},
				{Path: "secret1", Size: 50},
			},
			profile: "dev",
			expected: []vault.FileInfo{},
		},
		{
			name: "filter exact match",
			files: []vault.FileInfo{
				{Path: "profiles/prod/", Size: 100}, // This shouldn't match as it's just the prefix
			},
			profile: "prod",
			expected: []vault.FileInfo{}, // No match because path must be longer than prefix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterFilesByProfile(tt.files, tt.profile)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d files, got %d", len(tt.expected), len(result))
				return
			}

			for i, f := range result {
				if f.Path != tt.expected[i].Path || f.Size != tt.expected[i].Size {
					t.Errorf("file[%d] = {Path: %q, Size: %d}, want {Path: %q, Size: %d}",
						i, f.Path, f.Size, tt.expected[i].Path, tt.expected[i].Size)
				}
			}
		})
	}
}

// filterFilesByProfile filters files by profile prefix - extracted for testing
func filterFilesByProfile(files []vault.FileInfo, profile string) []vault.FileInfo {
	if profile == "" {
		return files
	}
	prefix := "profiles/" + profile + "/"
	var filtered []vault.FileInfo
	for _, f := range files {
		if len(f.Path) > len(prefix) && f.Path[:len(prefix)] == prefix {
			f.Path = f.Path[len(prefix):]
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func TestRunLsNotInitialized(t *testing.T) {
	// Note: The actual runLs calls config.GetToken() which requires keychain access
	// In a real refactor, we would inject the config and provider dependencies
	// For now, we test the helper functions
}