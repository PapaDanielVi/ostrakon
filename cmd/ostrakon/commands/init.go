package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Ostrakon and create the remote vault",
	Long: `Initialize Ostrakon by:
1. Storing your GitHub PAT securely in the OS keychain
2. Setting a master password for encryption
3. Creating the private 'ostrakon-vault' repository on GitHub`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Check if already initialized
	if pat, err := config.GetPAT(); err == nil && pat != "" {
		fmt.Println("Ostrakon is already initialized. Use 'ostrakon shred --all' to reset.")
		return nil
	}

	// Prompt for GitHub PAT
	fmt.Print("Enter your GitHub Personal Access Token (with repo scope): ")
	pat, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read PAT: %w", err)
	}
	pat = strings.TrimSpace(pat)
	if pat == "" {
		return fmt.Errorf("PAT cannot be empty")
	}

	// Create GitHub client to validate PAT
	client, err := github.NewClient(pat)
	if err != nil {
		return fmt.Errorf("invalid PAT: %w", err)
	}

	// Store PAT in keychain
	if err := config.StorePAT(pat); err != nil {
		return fmt.Errorf("failed to store PAT: %w", err)
	}
	fmt.Println("GitHub PAT stored securely in keychain")

	// Prompt for master password
	fmt.Print("Set a master password for encryption: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	fmt.Print("Confirm master password: ")
	passwordConfirm, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password != passwordConfirm {
		return fmt.Errorf("passwords do not match")
	}

	// Store password hash for validation
	hash := crypto.HashPassword(password)
	if err := config.StorePasswordHash(hash); err != nil {
		return fmt.Errorf("failed to store password hash: %w", err)
	}
	fmt.Println("Master password configured")

	// Create vault repository
	fmt.Println("Creating ostrakon-vault repository...")
	if err := client.EnsureVault(ctx); err != nil {
		// Clean up stored credentials on failure
		config.DeletePAT()
		config.DeletePasswordHash()
		return fmt.Errorf("failed to create vault: %w", err)
	}

	fmt.Println("Ostrakon initialized successfully!")
	fmt.Println("Your vault is ready at: https://github.com/" + client.Owner() + "/ostrakon-vault")
	return nil
}

func readPassword() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(password, "\n"), nil
}
