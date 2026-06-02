// Package gitlab provides a client for interacting with GitLab repositories as vaults.
package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PapaDanielVi/ostrakon/pkg/vault"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// Client wraps the GitLab client and provides vault operations.
type Client struct {
	glClient  *gitlab.Client
	projectID any
}

// NewClient creates a new GitLab client using the provided token and project ID.
// ProjectID can be either a numeric ID (int) or namespace/project string.
func NewClient(token string, projectID any) (*Client, error) {
	if token == "" {
		return nil, errors.New("token cannot be empty")
	}
	if projectID == nil {
		return nil, errors.New("project ID is required")
	}

	glClient, err := gitlab.NewClient(token, gitlab.WithBaseURL("https://gitlab.com/api/v4"))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &Client{
		glClient:  glClient,
		projectID: projectID,
	}, nil
}

// NewClientFromURL creates a new GitLab client from a repository URL and token.
// Supports HTTPS URLs like: https://gitlab.com/namespace/project.
// Supports short format: namespace/project.
func NewClientFromURL(repoURL, token string) (*Client, error) {
	projectID, err := ParseRepoURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repo URL: %w", err)
	}
	return NewClient(token, projectID)
}

// ParseRepoURL parses a GitLab repository URL and returns the project ID.
// GitLab accepts both numeric IDs and string IDs in "namespace/project" format.
func ParseRepoURL(repoURL string) (any, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return nil, errors.New("repo URL cannot be empty")
	}

	// Handle HTTPS URLs
	if strings.HasPrefix(repoURL, "https://") || strings.HasPrefix(repoURL, "http://") {
		repoURL = strings.TrimSuffix(repoURL, ".git")
		repoURL = strings.TrimSuffix(repoURL, "/")

		// Extract namespace/project from URL
		// https://gitlab.com/namespace/project
		parts := strings.Split(repoURL, "/")
		var project string
		for i, part := range parts {
			if part == "gitlab.com" || strings.HasSuffix(part, "gitlab") {
				if i+2 < len(parts) {
					project = parts[i+1] + "/" + parts[i+2]
					break
				}
			}
		}
		if project == "" {
			return nil, errors.New("could not extract project from URL")
		}
		return project, nil
	}

	// Handle short format: namespace/project
	if strings.Contains(repoURL, "/") {
		return repoURL, nil
	}

	// Check if it's a numeric ID
	if isNumeric(repoURL) {
		var id int
		if _, err := fmt.Sscanf(repoURL, "%d", &id); err == nil {
			return id, nil
		}
	}

	return nil, errors.New("invalid GitLab repo URL format")
}

// isNumeric checks if a string is a numeric value.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// CheckConnectivity verifies that the token has access to the project.
func (c *Client) CheckConnectivity(ctx context.Context) error {
	_, resp, err := c.glClient.Projects.GetProject(c.projectID, nil, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return errors.New("project not found or access denied")
		}
		return fmt.Errorf("failed to connect to project: %w", err)
	}
	return nil
}

var (
	branch = "main"
)

// getFileOptions returns common options for file operations.
func getFileOptions() *gitlab.GetFileOptions {
	return &gitlab.GetFileOptions{Ref: &branch}
}

// UploadFile uploads or updates a file in the vault.
func (c *Client) UploadFile(ctx context.Context, path string, content []byte, message string) error {
	fullPath := fmt.Sprintf("contents/%s", path)

	contentStr := string(content)

	// Check if file exists to get LastCommitID for update.
	fileContent, resp, err := c.glClient.RepositoryFiles.GetFile(c.projectID, fullPath, getFileOptions(), gitlab.WithContext(ctx))
	if err == nil && fileContent != nil {
		// Update existing file.
		lastCommitID := fileContent.LastCommitID
		_, _, err = c.glClient.RepositoryFiles.UpdateFile(c.projectID, fullPath, &gitlab.UpdateFileOptions{
			Branch:        &branch,
			Content:       &contentStr,
			CommitMessage: &message,
			LastCommitID:  &lastCommitID,
		}, gitlab.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to update file: %w", err)
		}
		return nil
	}
	if resp == nil || resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to check existing file: %w", err)
	}

	// Create new file.
	_, _, err = c.glClient.RepositoryFiles.CreateFile(c.projectID, fullPath, &gitlab.CreateFileOptions{
		Branch:        &branch,
		Content:       &contentStr,
		CommitMessage: &message,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// DownloadFile downloads a file from the vault.
func (c *Client) DownloadFile(ctx context.Context, path string) ([]byte, error) {
	fullPath := fmt.Sprintf("contents/%s", path)

	fileContent, _, err := c.glClient.RepositoryFiles.GetFile(c.projectID, fullPath, getFileOptions(), gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return []byte(fileContent.Content), nil
}

// DeleteFile deletes a file from the vault.
func (c *Client) DeleteFile(ctx context.Context, path, sha, message string) error {
	fullPath := fmt.Sprintf("contents/%s", path)

	// Get the LastCommitID if not provided (GitLab requires it for delete)
	if sha == "" {
		fileContent, _, err := c.glClient.RepositoryFiles.GetFile(c.projectID, fullPath, getFileOptions(), gitlab.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to get file: %w", err)
		}
		sha = fileContent.LastCommitID
	}

	_, err := c.glClient.RepositoryFiles.DeleteFile(c.projectID, fullPath, &gitlab.DeleteFileOptions{
		Branch:        &branch,
		LastCommitID:  &sha,
		CommitMessage: &message,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListFiles lists all files in the vault.
func (c *Client) ListFiles(ctx context.Context) ([]vault.FileInfo, error) {
	branch := "main"
	path := "contents/"
	recursive := true

	// Try main branch first, then master
	for _, br := range []string{branch, "master"} {
		tree, resp, err := c.glClient.Repositories.ListTree(c.projectID, &gitlab.ListTreeOptions{
			Ref:       &br,
			Path:      &path,
			Recursive: &recursive,
		}, gitlab.WithContext(ctx))
		if err == nil && len(tree) > 0 {
			var files []vault.FileInfo
			for _, entry := range tree {
				if entry.Type == "blob" {
					p := entry.Path
					files = append(files, vault.FileInfo{
						Path:      strings.TrimPrefix(p, "contents/"),
						SHA:       entry.ID, // TreeNode uses ID for blob identifier
						Size:      0,        // TreeNode doesn't provide size
						UpdatedAt: time.Now(),
					})
				}
			}
			return files, nil
		}
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get file tree: %w", err)
		}
	}

	return nil, errors.New("failed to get file tree: no valid branch found")
}

// GetFileSHA returns the SHA (BlobID) of a file for update operations.
func (c *Client) GetFileSHA(ctx context.Context, path string) (string, error) {
	fullPath := fmt.Sprintf("contents/%s", path)

	fileContent, resp, err := c.glClient.RepositoryFiles.GetFile(c.projectID, fullPath, getFileOptions(), gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to get file: %w", err)
	}

	return fileContent.BlobID, nil
}

// Owner returns the project path with namespace (e.g., "namespace/project").
func (c *Client) Owner() string {
	if s, ok := c.projectID.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", c.projectID)
}

// Repo returns the project identifier.
func (c *Client) Repo() string {
	if s, ok := c.projectID.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", c.projectID)
}

// RepoURL returns the full project URL.
func (c *Client) RepoURL() string {
	return fmt.Sprintf("https://gitlab.com/%v", c.projectID)
}

// ListCommits returns commits that affect the specified path.
func (c *Client) ListCommits(ctx context.Context, path string) ([]vault.CommitInfo, error) {
	commits, _, err := c.glClient.Commits.ListCommits(c.projectID, &gitlab.ListCommitsOptions{
		Path: &path,
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	var result []vault.CommitInfo
	for _, commit := range commits {
		var date time.Time
		if commit.CommittedDate != nil {
			date = *commit.CommittedDate
		}
		result = append(result, vault.CommitInfo{
			SHA:  commit.ID,
			Date: date,
		})
	}

	return result, nil
}

// ResetBranchToCommit resets the branch to point to a specific commit SHA.
// This effectively wipes all history after that commit for the repository.
func (c *Client) ResetBranchToCommit(_ context.Context, _, _ string) error {
	// GitLab doesn't support server-side branch reset via API in a simple way
	// We would need to use git push --force workflow for this operation
	// For now, we return an error indicating this needs custom handling
	return errors.New("hard reset via API requires git push --force workflow. Use git clone/push instead")
}
