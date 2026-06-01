// Package github provides a client for interacting with GitHub repositories as vaults.
package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PapaDanielVi/ostrakon/pkg/vault"
	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub client and provides vault operations.
type Client struct {
	ghClient *github.Client
	owner    string
	repo     string
}

// NewClient creates a new GitHub client using the provided token and repo.
func NewClient(token, repoOwner, repoName string) (*Client, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}
	if repoOwner == "" || repoName == "" {
		return nil, errors.New("owner and repo name are required")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)

	return &Client{
		ghClient: ghClient,
		owner:    repoOwner,
		repo:     repoName,
	}, nil
}

// NewClientFromURL creates a new GitHub client from a repository URL and token.
// Supports HTTPS URLs like: https://github.com/owner/repo.
// Supports SSH URLs like: git@github.com:owner/repo.git.
func NewClientFromURL(repoURL, token string) (*Client, error) {
	owner, repo, err := ParseRepoURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repo URL: %w", err)
	}
	return NewClient(token, owner, repo)
}

// ParseRepoURL parses a GitHub repository URL and returns owner and repo name.
// Supports HTTPS URLs: https://github.com/owner/repo.
// Supports SSH URLs: git@github.com:owner/repo.git.
// Supports short URLs: owner/repo.
func ParseRepoURL(repoURL string) (string, string, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return "", "", errors.New("repo URL cannot be empty")
	}

	// Handle SSH format: git@github.com:owner/repo.git
	if strings.HasPrefix(repoURL, "git@") {
		parts := strings.Split(repoURL, ":")
		if len(parts) != 2 {
			return "", "", errors.New("invalid SSH repo URL format")
		}
		pathPart := parts[1] // owner/repo.git
		pathPart = strings.TrimSuffix(pathPart, ".git")
		parts = strings.Split(pathPart, "/")
		if len(parts) != 2 {
			return "", "", errors.New("invalid SSH repo URL format")
		}
		return parts[0], parts[1], nil
	}

	// Handle HTTPS/short format
	// Remove trailing .git if present
	repoURL = strings.TrimSuffix(repoURL, ".git")
	// Remove trailing slash
	repoURL = strings.TrimSuffix(repoURL, "/")

	// If it's a full HTTPS URL
	if strings.HasPrefix(repoURL, "https://") || strings.HasPrefix(repoURL, "http://") {
		parts := strings.Split(repoURL, "/")
		// Find the owner and repo in the URL path
		// https://github.com/owner/repo -> [https:, "", github.com, owner, repo]
		if len(parts) < 4 {
			return "", "", errors.New("invalid HTTPS repo URL format")
		}
		owner := parts[3]
		repo := parts[4]
		if owner == "" || repo == "" {
			return "", "", errors.New("invalid HTTPS repo URL format")
		}
		return owner, repo, nil
	}

	// Handle short format: owner/repo
	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid short repo format, expected owner/repo")
	}
	return parts[0], parts[1], nil
}

// CheckConnectivity verifies that the token has access to the repository.
func (c *Client) CheckConnectivity(ctx context.Context) error {
	_, resp, err := c.ghClient.Repositories.Get(ctx, c.owner, c.repo)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("repository not found or access denied: %s/%s", c.owner, c.repo)
		}
		return fmt.Errorf("failed to connect to repository: %w", err)
	}
	return nil
}

// GetRepository returns the repository info.
func (c *Client) GetRepository(ctx context.Context) (*github.Repository, error) {
	repo, _, err := c.ghClient.Repositories.Get(ctx, c.owner, c.repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return repo, nil
}

// UploadFile uploads or updates a file in the vault.
func (c *Client) UploadFile(ctx context.Context, path string, content []byte, message string) error {
	fullPath := fmt.Sprintf("contents/%s", path)

	// Check if file exists to get SHA
	sha := ""
	fileContent, _, resp, err := c.ghClient.Repositories.GetContents(ctx, c.owner, c.repo, fullPath, nil)
	if err == nil && fileContent != nil {
		sha = fileContent.GetSHA()
	} else if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to check existing file: %w", err)
	}

	committerName := "Ostrakon"
	committerEmail := "ostrakon@cli"
	fileOpts := &github.RepositoryContentFileOptions{
		Message:   &message,
		Content:   content,
		Committer: &github.CommitAuthor{Name: &committerName, Email: &committerEmail},
	}

	if sha != "" {
		fileOpts.SHA = &sha
	}

	_, _, err = c.ghClient.Repositories.UpdateFile(ctx, c.owner, c.repo, fullPath, fileOpts)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from the vault.
func (c *Client) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	fullPath := fmt.Sprintf("contents/%s", path)

	fileContent, _, _, err := c.ghClient.Repositories.GetContents(ctx, c.owner, c.repo, fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	return []byte(content), nil
}

// DeleteFile deletes a file from the vault.
func (c *Client) DeleteFile(ctx context.Context, path, sha, message string) error {
	fullPath := fmt.Sprintf("contents/%s", path)

	committerName := "Ostrakon"
	committerEmail := "ostrakon@cli"
	_, _, err := c.ghClient.Repositories.DeleteFile(ctx, c.owner, c.repo, fullPath, &github.RepositoryContentFileOptions{
		Message:   &message,
		SHA:       &sha,
		Committer: &github.CommitAuthor{Name: &committerName, Email: &committerEmail},
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListFiles lists all files in the vault using the Tree API.
func (c *Client) ListFiles(ctx context.Context) ([]vault.FileInfo, error) {
	// Try main branch first
	tree, _, err := c.ghClient.Git.GetTree(ctx, c.owner, c.repo, "main", true)
	if err != nil {
		// Try 'master' branch if main doesn't exist
		tree, _, err = c.ghClient.Git.GetTree(ctx, c.owner, c.repo, "master", true)
		if err != nil {
			return nil, fmt.Errorf("failed to get file tree: %w", err)
		}
	}

	var files []vault.FileInfo
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			// Skip the contents/ prefix
			path := entry.GetPath()
			if len(path) > 9 && path[:9] == "contents/" {
				path = path[9:]
			}
			files = append(files, vault.FileInfo{
				Path:      path,
				SHA:       entry.GetSHA(),
				Size:      entry.GetSize(),
				UpdatedAt: time.Now(), // Trees API doesn't provide modification time
			})
		}
	}

	return files, nil
}

// GetFileSHA returns the SHA of a file for update operations.
func (c *Client) GetFileSHA(ctx context.Context, path string) (string, error) {
	fullPath := fmt.Sprintf("contents/%s", path)

	fileContent, _, resp, err := c.ghClient.Repositories.GetContents(ctx, c.owner, c.repo, fullPath, nil)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	return fileContent.GetSHA(), nil
}

// Owner returns the authenticated user's login (repo owner).
func (c *Client) Owner() string {
	return c.owner
}

// Repo returns the repository name.
func (c *Client) Repo() string {
	return c.repo
}

// RepoURL returns the full repository URL.
func (c *Client) RepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", c.owner, c.repo)
}

// ListCommits returns commits that affect the specified path.
func (c *Client) ListCommits(ctx context.Context, path string) ([]vault.CommitInfo, error) {
	opts := &github.CommitsListOptions{
		Path: path,
	}

	commits, _, err := c.ghClient.Repositories.ListCommits(ctx, c.owner, c.repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	var result []vault.CommitInfo
	for _, commit := range commits {
		var date time.Time
		if commit.Commit.Committer != nil && commit.Commit.Committer.Date != nil {
			date = commit.Commit.Committer.Date.Time
		}
		result = append(result, vault.CommitInfo{
			SHA:  commit.GetSHA(),
			Date: date,
		})
	}

	return result, nil
}

// ResetBranchToCommit resets the branch to point to a specific commit SHA.
// This effectively wipes all history after that commit for the repository.
func (c *Client) ResetBranchToCommit(ctx context.Context, branch, sha string) error {
	// GitHub API doesn't support server-side hard reset directly
	// We would need to use git references API to update the branch ref
	// However, this is a destructive operation and GitHub's API has limitations
	// For now, we return an error indicating this needs custom handling
	return fmt.Errorf("hard reset via API requires git push --force workflow. Use git clone/push instead")
}

// ReadFileFromStdin reads content from standard input.
func ReadFileFromStdin() ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, errors.New("no data piped to stdin")
	}
	return io.ReadAll(os.Stdin)
}