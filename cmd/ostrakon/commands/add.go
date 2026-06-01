package commands

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/provider"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path> [-n name]",
	Short: "Add a file to the vault",
	Long: `Encrypt and upload a file to the vault.
Reads from stdin if data is piped.

Paths with slashes are supported for hierarchical organization:
  ostrakon add prod/db/password
  ostrakon add staging/api/key`,
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
	ctx := cmd.Context()

	// Get vault client from provider factory
	client, err := provider.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Read content
	var content []byte
	isPiped := false
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("failed to check stdin: %w", err)
	}
	switch {
	case (stat.Mode() & os.ModeCharDevice) == 0:
		// Data is piped
		isPiped = true
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	case len(args) > 0:
		// Read from file
		content, err = os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	default:
		return errors.New("no input provided. Pipe data or provide a file path")
	}

	// Determine vault path
	name := fileName
	if name == "" {
		if len(args) > 0 {
			name = args[0]
		} else {
			return errors.New("no name specified. Use -n flag or provide a file path")
		}
	}

	vaultPath := name
	if addProfile != "" {
		vaultPath = fmt.Sprintf("profiles/%s/%s", addProfile, name)
	}

	// Get master password from keyring (no prompt for write operations)
	password, err := getPassword()
	if err != nil {
		// If not in keyring and data was piped, prompt after reading stdin
		if isPiped {
			password, err = readPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		} else {
			return err
		}
	}

	// Validate password
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}
	if !crypto.ValidatePassword(password, hash) {
		return errors.New("invalid password")
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