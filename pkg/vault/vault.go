// Package vault provides the interfaces for vault provider.
package vault

import (
	"context"
	"time"
)

// Provider defines the interface for vault operations.
type Provider interface {
	CheckConnectivity(ctx context.Context) error
	UploadFile(ctx context.Context, path string, content []byte, message string) error
	DownloadFile(ctx context.Context, path string) ([]byte, error)
	DeleteFile(ctx context.Context, path, sha, message string) error
	ListFiles(ctx context.Context) ([]FileInfo, error)
	GetFileSHA(ctx context.Context, path string) (string, error)
	Owner() string
	Repo() string
	RepoURL() string
}

// FileInfo represents information about a file in the vault.
type FileInfo struct {
	Path      string
	SHA       string
	Size      int
	UpdatedAt time.Time
}

// CommitInfo represents information about a commit for history operations.
type CommitInfo struct {
	SHA  string
	Date time.Time
}

// HistoryProvider extends Provider with commit history operations.
// This is used for the --hard flag to wipe commit history.
type HistoryProvider interface {
	Provider
	ListCommits(ctx context.Context, path string) ([]CommitInfo, error)
	ResetBranchToCommit(ctx context.Context, branch, sha string) error
}