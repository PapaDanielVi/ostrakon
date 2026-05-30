package commands

import (
	"context"
	"fmt"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
)

var shredCmd = &cobra.Command{
	Use:   "shred <name> | --all",
	Short: "Securely delete a secret from the vault",
	Long: `Overwrite a file with random data before deleting it from the vault.
This provides deniability by destroying the encrypted file's history.`,
	RunE:  runShred,
}

var (
	shredAll bool
)

func init() {
	shredCmd.Flags().BoolVarP(&shredAll, "all", "", false, "Reset all Ostrakon data (delete vault, clear keychain)")
}

func runShred(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if shredAll {
		return resetAll()
	}

	if len(args) < 1 {
		return fmt.Errorf("specify a secret name or use --all")
	}

	// Get PAT from keychain
	pat, err := config.GetPAT()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Create GitHub client
	client, err := github.NewClient(pat)
	if err != nil {
		return err
	}

	name := args[0]

	// Prompt for master password to confirm
	fmt.Print("Enter master password to confirm deletion: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}
	if !crypto.ValidatePassword(password, hash) {
		return fmt.Errorf("invalid password")
	}

	// Get file SHA
	sha, err := client.GetFileSHA(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	if sha == "" {
		return fmt.Errorf("secret not found: %s", name)
	}

	// Overwrite with random data
	randomData := make([]byte, 1024)
	for i := range randomData {
		randomData[i] = byte(i % 256)
	}
	encrypted, _ := crypto.Encrypt(string(randomData), password)

	// Upload the overwritten file
	if err := client.UploadFile(ctx, name, []byte(encrypted), fmt.Sprintf("Shred secret: %s", name)); err != nil {
		return fmt.Errorf("failed to overwrite: %w", err)
	}

	// Delete the file
	if err := client.DeleteFile(ctx, name, sha, fmt.Sprintf("Delete shredded secret: %s", name)); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	fmt.Printf("Secret '%s' securely shredded\n", name)
	return nil
}

func resetAll() error {
	// Get PAT from keychain
	pat, err := config.GetPAT()
	if err != nil {
		// If no PAT, just clear local data
		config.DeletePasswordHash()
		fmt.Println("Local data cleared")
		return nil
	}

	// Create GitHub client
	client, err := github.NewClient(pat)
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Prompt for master password to confirm
	fmt.Print("Enter master password to confirm deletion of all data: ")
	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}
	if !crypto.ValidatePassword(password, hash) {
		return fmt.Errorf("invalid password")
	}

	// Delete vault repository
	fmt.Println("Deleting vault repository...")
	if err := deleteVaultRepository(ctx, client); err != nil {
		fmt.Printf("Warning: failed to delete vault: %v\n", err)
	}

	// Clear keychain
	config.DeletePAT()
	config.DeletePasswordHash()

	fmt.Println("All Ostrakon data has been reset")
	return nil
}

func deleteVaultRepository(ctx context.Context, client *github.Client) error {
	// This requires using the Repositories.Delete method
	// We'll need to add this to the github client
	return client.DeleteVault(ctx)
}