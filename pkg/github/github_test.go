package github

import (
	"context"
	"testing"
)

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "HTTPS URL",
			repoURL:   "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with .git",
			repoURL:   "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with trailing slash",
			repoURL:   "https://github.com/owner/repo/",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL trailing slash and .git - note: .git suffix remains",
			repoURL:   "https://github.com/owner/repo.git/",
			wantOwner: "owner",
			wantRepo:  "repo.git",
			wantErr:   false,
		},
		{
			name:      "SSH URL",
			repoURL:   "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH URL without .git",
			repoURL:   "git@github.com:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "short format",
			repoURL:   "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "short format with whitespace",
			repoURL:   "  owner/repo  ",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:    "empty URL",
			repoURL: "",
			wantErr: true,
		},
		{
			name:    "invalid SSH format missing colon",
			repoURL: "git@github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "invalid SSH format multiple colons",
			repoURL: "git@github.com::owner/repo",
			wantErr: true,
		},
		{
			name:    "invalid short format missing repo",
			repoURL: "owner",
			wantErr: true,
		},
		{
			name:    "invalid short format multiple slashes",
			repoURL: "owner/repo/extra",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepoURL(tt.repoURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRepoURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("ParseRepoURL() owner = %v, want %v", owner, tt.wantOwner)
				}
				if repo != tt.wantRepo {
					t.Errorf("ParseRepoURL() repo = %v, want %v", repo, tt.wantRepo)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		owner   string
		repo    string
		wantErr bool
	}{
		{
			name:    "valid client",
			token:   "ghp_testtoken",
			owner:   "owner",
			repo:    "repo",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			owner:   "owner",
			repo:    "repo",
			wantErr: true,
		},
		{
			name:    "empty owner",
			token:   "ghp_testtoken",
			owner:   "",
			repo:    "repo",
			wantErr: true,
		},
		{
			name:    "empty repo",
			token:   "ghp_testtoken",
			owner:   "owner",
			repo:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.token, tt.owner, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if client.Owner() != tt.owner {
					t.Errorf("NewClient().Owner() = %v, want %v", client.Owner(), tt.owner)
				}
				if client.Repo() != tt.repo {
					t.Errorf("NewClient().Repo() = %v, want %v", client.Repo(), tt.repo)
				}

				expectedURL := "https://github.com/" + tt.owner + "/" + tt.repo
				if client.RepoURL() != expectedURL {
					t.Errorf("NewClient().RepoURL() = %v, want %v", client.RepoURL(), expectedURL)
				}
			}
		})
	}
}

func TestNewClientFromURL(t *testing.T) {
	tests := []struct {
		name      string
		repoURL   string
		token     string
		wantErr   bool
		wantOwner string
		wantRepo  string
	}{
		{
			name:      "valid HTTPS URL",
			repoURL:   "https://github.com/owner/repo",
			token:     "ghp_testtoken",
			wantErr:   false,
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:    "invalid URL",
			repoURL: "invalid",
			token:   "ghp_testtoken",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientFromURL(tt.repoURL, tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewClientFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if client.Owner() != tt.wantOwner {
					t.Errorf("NewClientFromURL().Owner() = %v, want %v", client.Owner(), tt.wantOwner)
				}
				if client.Repo() != tt.wantRepo {
					t.Errorf("NewClientFromURL().Repo() = %v, want %v", client.Repo(), tt.wantRepo)
				}
			}
		})
	}
}

func TestClientOwnerRepo(t *testing.T) {
	client, err := NewClient("token", "testowner", "testrepo")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.Owner() != "testowner" {
		t.Errorf("Owner() = %v, want testowner", client.Owner())
	}
	if client.Repo() != "testrepo" {
		t.Errorf("Repo() = %v, want testrepo", client.Repo())
	}
	if client.RepoURL() != "https://github.com/testowner/testrepo" {
		t.Errorf("RepoURL() = %v, want https://github.com/testowner/testrepo", client.RepoURL())
	}
}

func TestReadFileFromStdin(t *testing.T) {
	_, err := ReadFileFromStdin()
	if err == nil {
		t.Error("expected error when stdin is not piped")
	}
}

func TestClientCheckConnectivityUninitialized(t *testing.T) {
	client, err := NewClient("test_token", "owner", "repo")
	if err != nil {
		t.Fatalf("NewClient failed unexpectedly: %v", err)
	}

	err = client.CheckConnectivity(context.Background())
	if err == nil {
		t.Error("expected error for invalid token/repo")
	}
}