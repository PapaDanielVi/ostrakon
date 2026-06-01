// Package provider provides a factory for creating vault provider clients.
package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/PapaDanielVi/ostrakon/pkg/gitlab"
	"github.com/PapaDanielVi/ostrakon/pkg/vault"
)

// NewClient creates a vault provider client based on stored configuration.
// It reads the provider type from the keychain and returns the appropriate client.
func NewClient(ctx context.Context) (vault.Provider, error) {
	providerType, err := config.GetProviderType()
	if err != nil {
		return nil, fmt.Errorf("failed to get provider type: %w", err)
	}

	token, err := config.GetToken()
	if err != nil {
		return nil, fmt.Errorf("not initialized: %w", err)
	}

	switch providerType {
	case config.ProviderGitLab:
		// For GitLab, we store projectID in RepoOwnerKey and project name in RepoNameKey
		// or we can use the URL format
		projectID, err := config.GetRepoURL()
		if err != nil {
			return nil, fmt.Errorf("not initialized: %w", err)
		}
		// Parse the URL to get projectID
		parsedProjectID, err := gitlab.ParseRepoURL(projectID)
		if err != nil {
			return nil, fmt.Errorf("invalid GitLab project URL: %w", err)
		}
		return gitlab.NewClient(token, parsedProjectID)
	case config.ProviderGitHub:
		fallthrough
	default:
		owner, err := config.GetRepoOwner()
		if err != nil {
			return nil, fmt.Errorf("not initialized: %w", err)
		}
		repo, err := config.GetRepoName()
		if err != nil {
			return nil, fmt.Errorf("not initialized: %w", err)
		}
		return github.NewClient(token, owner, repo)
	}
}

// NewClientFromURL creates a vault provider client from a repository URL and token.
// This is used during initialization to determine the correct provider.
func NewClientFromURL(repoURL, token string) (vault.Provider, error) {
	if repoURL == "" {
		return nil, errors.New("repo URL cannot be empty")
	}

	// Detect provider from URL
	if strings.Contains(repoURL, "gitlab") {
		return gitlab.NewClientFromURL(repoURL, token)
	}

	// Default to GitHub
	return github.NewClientFromURL(repoURL, token)
}

// StoreProviderInfo stores the provider type and repository info in the keychain.
func StoreProviderInfo(providerType, repoURL string) error {
	if providerType == "" {
		return errors.New("provider type cannot be empty")
	}

	if err := config.StoreProviderType(providerType); err != nil {
		return fmt.Errorf("failed to store provider type: %w", err)
	}

	// Parse and store repo info based on provider type
	switch providerType {
	case config.ProviderGitLab:
		// For GitLab, we store the full URL since we use namespace/project format
		projectID, err := gitlab.ParseRepoURL(repoURL)
		if err != nil {
			return fmt.Errorf("invalid GitLab project URL: %w", err)
		}
		// Extract namespace/project for storage
		if s, ok := projectID.(string); ok {
			parts := strings.SplitN(s, "/", 2)
			if len(parts) == 2 {
				if err := config.StoreRepoInfo(repoURL, parts[0], parts[1]); err != nil {
					return fmt.Errorf("failed to store repo info: %w", err)
				}
			}
		}
	case config.ProviderGitHub:
		fallthrough
	default:
		// Use GitHub parsing
		owner, repo, err := github.ParseRepoURL(repoURL)
		if err != nil {
			return fmt.Errorf("invalid GitHub repo URL: %w", err)
		}
		if err := config.StoreRepoInfo(repoURL, owner, repo); err != nil {
			return fmt.Errorf("failed to store repo info: %w", err)
		}
	}

	return nil
}