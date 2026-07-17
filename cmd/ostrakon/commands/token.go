package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/provider"
	"github.com/spf13/cobra"
)

var updateTokenCmd = &cobra.Command{
	Use:   "update-token",
	Short: "Update the stored access token",
	Long: `Replace the access token stored in the OS keychain without touching your
secrets or master password.

Use this when your token has expired or been rotated. The new token is verified
against your repository before it replaces the old one, so an invalid token
leaves your existing configuration untouched.`,
	RunE: runUpdateToken,
}

func runUpdateToken(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Ensure the vault is already initialized before updating the token.
	if _, err := config.GetToken(); err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Prompt for the new token without echoing it.
	token, err := readPasswordPrompt("Enter your new access token with contents:read and contents:write permissions: ")
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token cannot be empty")
	}

	// Verify the new token against the repository before persisting it.
	fmt.Println("\nVerifying new token...")
	client, err := provider.NewClientWithToken(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	if err := client.CheckConnectivity(ctx); err != nil {
		return fmt.Errorf("connectivity check failed: %w", err)
	}
	fmt.Println("  ✓ New token is valid")

	// Persist the verified token, replacing the old one.
	if err := config.StoreToken(token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}
	fmt.Println("  ✓ Access token updated in keychain")

	return nil
}
