package commands

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <file> [-n name]",
	Short: "Add a file to the vault",
	Long: `Encrypt and upload a file to the vault.
Reads from stdin if data is piped.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

var (
	fileName   string
	addProfile string
)

func init() {
	addCmd.Flags().StringVarP(&fileName, "name", "n", "", "Name for the file in the vault")
	addCmd.Flags().StringVarP(&addProfile, "profile", "p", "", "Profile/namespace for the file")
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

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

	// Check if vault exists
	exists, err := client.VaultExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("vault does not exist. Run 'ostrakon init' first")
	}

	// Read content
	var content []byte
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is piped
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	} else if len(args) > 0 {
		// Read from file
		content, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	} else {
		return fmt.Errorf("no input provided. Pipe data or provide a file path")
	}

	// Determine vault path
	name := fileName
	if name == "" {
		if len(args) > 0 {
			name = args[0]
		} else {
			return fmt.Errorf("no name specified. Use -n flag or provide a file path")
		}
	}

	vaultPath := name
	if addProfile != "" {
		vaultPath = fmt.Sprintf("profiles/%s/%s", addProfile, name)
	}

	// Prompt for master password
	fmt.Print("Enter master password: ")
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

	// Encrypt content
	encrypted, err := crypto.Encrypt(string(content), password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Upload to vault
	message := fmt.Sprintf("Add/update secret: %s", vaultPath)
	if err := client.UploadFile(ctx, vaultPath, []byte(encrypted), message); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("Secret '%s' uploaded to vault\n", vaultPath)
	return nil
}
