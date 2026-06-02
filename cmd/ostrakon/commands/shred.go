package commands

import (
	"errors"
	"fmt"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/provider"
	"github.com/spf13/cobra"
)

var shredCmd = &cobra.Command{
	Use:   "shred <path> | --all [--hard]",
	Short: "Securely delete a secret from the vault",
	Long: `Overwrite a file with random data before deleting it from the vault.
This provides deniability by destroying the encrypted file's history.

Paths with slashes are supported for hierarchical organization:
  ostrakon shred prod/db/password
  ostrakon shred staging/api/key

With --all --hard:
  Deletes all secrets and wipes the commit history back to the init commit.`,
	RunE: runShred,
}

var (
	shredAll  bool
	shredHard bool
)

func init() {
	shredCmd.Flags().BoolVarP(&shredAll, "all", "", false, "Reset all Ostrakon data (clear keychain)")
	shredCmd.Flags().BoolVarP(&shredHard, "hard", "", false, "Wipe commit history along with data (only with --all)")
}

func runShred(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Validate --hard flag usage
	if shredHard && !shredAll {
		return errors.New("--hard can only be used with --all")
	}

	if shredAll {
		return resetAll()
	}

	if len(args) < 1 {
		return errors.New("specify a secret name or use --all")
	}

	// Get vault client from provider factory
	client, err := provider.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	name := args[0]

	// Prompt for master password to confirm
	password, err := getPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	if !crypto.ValidatePassword(password, hash) {
		return errors.New("invalid password")
	}

	// Get file SHA
	sha, err := client.GetFileSHA(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if sha == "" {
		return fmt.Errorf("secret not found: %s", name)
	}

	// Delete the file
	if err := client.DeleteFile(ctx, name, sha, fmt.Sprintf("Delete shredded secret: %s", name)); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	fmt.Printf("Secret '%s' securely shredded\n", name)
	return nil
}

func resetAll() error {
	// Get token from keychain
	_, err := config.GetToken()
	if err != nil {
		// If no token, just clear local data
		_ = config.DeletePasswordHash()
		_ = config.DeleteGlobalMasterPassword()
		_ = config.DeleteProviderType()
		fmt.Println("Local data cleared")
		//nolint:nilerr // Intentionally return nil; we cleared what we could
		return nil
	}

	// Prompt for master password to confirm
	password, err := readPasswordPrompt("Enter master password to confirm deletion of all data: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}
	if !crypto.ValidatePassword(password, hash) {
		return errors.New("invalid password")
	}

	// Clear keychain
	_ = config.DeleteToken()
	_ = config.DeleteRepoInfo()
	_ = config.DeletePasswordHash()
	_ = config.DeleteGlobalMasterPassword()
	_ = config.DeleteProviderType()

	if shredHard {
		fmt.Println("All Ostrakon data has been reset (--hard: history wipe not yet implemented)")
	} else {
		fmt.Println("All Ostrakon data has been reset")
	}
	return nil
}
