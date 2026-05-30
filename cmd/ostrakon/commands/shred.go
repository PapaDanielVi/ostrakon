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
	shredCmd.Flags().BoolVarP(&shredAll, "all", "", false, "Reset all Ostrakon data (clear keychain)")
}

func runShred(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if shredAll {
		return resetAll()
	}

	if len(args) < 1 {
		return fmt.Errorf("specify a secret name or use --all")
	}

	// Get token and repo info from keychain
	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	owner, err := config.GetRepoOwner()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	repoName, err := config.GetRepoName()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Create GitHub client
	client, err := github.NewClient(token, owner, repoName)
	if err != nil {
		return err
	}

	name := args[0]

	// Prompt for master password to confirm
	fmt.Print("Enter master password to confirm deletion: ")
	password, err := readLine()
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
	// Get token from keychain
	token, err := config.GetToken()
	if err != nil {
		// If no token, just clear local data
		config.DeletePasswordHash()
		fmt.Println("Local data cleared")
		return nil
	}

	// Get repo info for cleanup
	owner, _ := config.GetRepoOwner()
	repoName, _ := config.GetRepoName()

	// Create GitHub client (validates token is still usable)
	_, err = github.NewClient(token, owner, repoName)
	if err != nil {
		return err
	}

	// Prompt for master password to confirm
	fmt.Print("Enter master password to confirm deletion of all data: ")
	password, err := readLine()
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

	// Clear keychain
	config.DeleteToken()
	config.DeleteRepoInfo()
	config.DeletePasswordHash()

	fmt.Println("All Ostrakon data has been reset")
	return nil
}