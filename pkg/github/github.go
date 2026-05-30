package github

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

const (
	VaultRepoName = "ostrakon-vault"
)

// Client wraps the GitHub client and provides vault operations
type Client struct {
	ghClient *github.Client
	owner    string
}

// NewClient creates a new GitHub client using the provided PAT
func NewClient(pat string) (*Client, error) {
	if pat == "" {
		return nil, errors.New("PAT cannot be empty")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: pat})
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)

	// Get the authenticated user
	user, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with GitHub: %w", err)
	}

	return &Client{
		ghClient: ghClient,
		owner:    user.GetLogin(),
	}, nil
}

// EnsureVault ensures the vault repository exists
func (c *Client) EnsureVault(ctx context.Context) error {
	// Check if repo exists
	_, resp, err := c.ghClient.Repositories.Get(ctx, c.owner, VaultRepoName)
	if err == nil {
		return nil // Repo exists
	}

	if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to check vault repository: %w", err)
	}

	// Create the repository
	repo := &github.Repository{
		Name:        github.String(VaultRepoName),
		Private:     github.Bool(true),
		Description: github.String("Ostrakon secure vault for encrypted secrets"),
		AutoInit:    github.Bool(true),
	}

	_, _, err = c.ghClient.Repositories.Create(ctx, "", repo)
	if err != nil {
		return fmt.Errorf("failed to create vault repository: %w", err)
	}

	return nil
}

// VaultExists checks if the vault repository exists
func (c *Client) VaultExists(ctx context.Context) (bool, error) {
	_, resp, err := c.ghClient.Repositories.Get(ctx, c.owner, VaultRepoName)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// UploadFile uploads or updates a file in the vault
func (c *Client) UploadFile(ctx context.Context, path string, content []byte, message string) error {
	fullPath := fmt.Sprintf("contents/%s", path)

	// Check if file exists to get SHA
	sha := ""
	fileContent, _, resp, err := c.ghClient.Repositories.GetContents(ctx, c.owner, VaultRepoName, fullPath, nil)
	if err == nil && fileContent != nil {
		sha = fileContent.GetSHA()
	} else if resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to check existing file: %w", err)
	}

	fileOpts := &github.RepositoryContentFileOptions{
		Message:   &message,
		Content:   content,
		Committer: &github.CommitAuthor{Name: github.String("Ostrakon"), Email: github.String("ostrakon@cli")},
	}

	if sha != "" {
		fileOpts.SHA = &sha
	}

	_, _, err = c.ghClient.Repositories.UpdateFile(ctx, c.owner, VaultRepoName, fullPath, fileOpts)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from the vault
func (c *Client) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	fullPath := fmt.Sprintf("contents/%s", path)

	fileContent, _, _, err := c.ghClient.Repositories.GetContents(ctx, c.owner, VaultRepoName, fullPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	return []byte(content), nil
}

// DeleteFile deletes a file from the vault
func (c *Client) DeleteFile(ctx context.Context, path, sha, message string) error {
	fullPath := fmt.Sprintf("contents/%s", path)

	_, _, err := c.ghClient.Repositories.DeleteFile(ctx, c.owner, VaultRepoName, fullPath, &github.RepositoryContentFileOptions{
		Message:   &message,
		SHA:       &sha,
		Committer: &github.CommitAuthor{Name: github.String("Ostrakon"), Email: github.String("ostrakon@cli")},
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListFiles lists all files in the vault using the Tree API
func (c *Client) ListFiles(ctx context.Context) ([]FileInfo, error) {
	// Use the Git Trees API for efficient listing
	tree, _, err := c.ghClient.Git.GetTree(ctx, c.owner, VaultRepoName, "main", true)
	if err != nil {
		// Try 'master' branch if main doesn't exist
		tree, _, err = c.ghClient.Git.GetTree(ctx, c.owner, VaultRepoName, "master", true)
		if err != nil {
			return nil, fmt.Errorf("failed to get file tree: %w", err)
		}
	}

	var files []FileInfo
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			// Skip the contents/ prefix
			path := entry.GetPath()
			if len(path) > 9 && path[:9] == "contents/" {
				path = path[9:]
			}
			files = append(files, FileInfo{
				Path:      path,
				SHA:       entry.GetSHA(),
				Size:      entry.GetSize(),
				UpdatedAt: time.Now(), // Trees API doesn't provide modification time
			})
		}
	}

	return files, nil
}

// GetFileSHA returns the SHA of a file for update operations
func (c *Client) GetFileSHA(ctx context.Context, path string) (string, error) {
	fullPath := fmt.Sprintf("contents/%s", path)

	fileContent, _, resp, err := c.ghClient.Repositories.GetContents(ctx, c.owner, VaultRepoName, fullPath, nil)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	return fileContent.GetSHA(), nil
}

// FileInfo represents information about a file in the vault
type FileInfo struct {
	Path      string
	SHA       string
	Size      int
	UpdatedAt time.Time
}

// Owner returns the authenticated user's login
func (c *Client) Owner() string {
	return c.owner
}

// DeleteVault deletes the vault repository
func (c *Client) DeleteVault(ctx context.Context) error {
	_, err := c.ghClient.Repositories.Delete(ctx, c.owner, VaultRepoName)
	if err != nil {
		return fmt.Errorf("failed to delete vault repository: %w", err)
	}
	return nil
}

// ReadFileFromStdin reads content from standard input
func ReadFileFromStdin() ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, errors.New("no data piped to stdin")
	}
	return io.ReadAll(os.Stdin)
}
