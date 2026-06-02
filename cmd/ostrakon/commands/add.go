package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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

// expandHomePath expands the tilde (~) to the user's home directory.
func expandHomePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return strings.Replace(path, "~", home, 1), nil
	}
	return path, nil
}

// extractLastTwoPathComponents extracts the last two components of a path.
// For example: /Users/mk/Documents/test.txt -> Documents/test.txt
// For relative paths like test.txt -> test.txt
func extractLastTwoPathComponents(path string) string {
	// Clean and normalize the path.
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))

	// Get the last two components.
	if len(parts) <= 1 {
		if len(parts) == 1 {
			return parts[0]
		}
		return path
	}
	return strings.Join(parts[len(parts)-2:], "/")
}

// resolveVaultName resolves the vault name from an argument path.
// It expands home directory and extracts the last two path components.
func resolveVaultName(argPath string) string {
	expandedPath, err := expandHomePath(argPath)
	if err != nil {
		expandedPath = argPath
	}
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		return argPath
	}
	return extractLastTwoPathComponents(absPath)
}

func runAdd(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	if Verbose {
		fmt.Println("Initializing vault client...")
	}
	// Get vault client from provider factory.
	client, err := provider.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Read content.
	var content []byte
	isPiped := false
	stat, err := os.Stdin.Stat()
	if err != nil {
		return fmt.Errorf("failed to check stdin: %w", err)
	}
	switch {
	case (stat.Mode() & os.ModeCharDevice) == 0:
		// Data is piped.
		if Verbose {
			fmt.Println("Reading content from stdin...")
		}
		isPiped = true
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
	case len(args) > 0:
		// Read from file.
		expandedPath, _ := expandHomePath(args[0])
		if Verbose {
			fmt.Printf("Reading file: %s\n", args[0])
		}
		content, err = os.ReadFile(expandedPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
	default:
		return errors.New("no input provided. Pipe data or provide a file path")
	}

	// Determine vault path.
	name := fileName
	if name == "" {
		if len(args) > 0 {
			name = resolveVaultName(args[0])
		} else {
			return errors.New("no name specified. Use -n flag or provide a file path")
		}
	}

	vaultPath := name
	if addProfile != "" {
		vaultPath = fmt.Sprintf("profiles/%s/%s", addProfile, name)
	}

	if Verbose {
		fmt.Printf("Resolved vault path: %s\n", vaultPath)
	}

	// Get master password from keyring (no prompt for write operations).
	password, err := getPassword()
	if err != nil {
		// If not in keyring and data was piped, prompt after reading stdin.
		if isPiped {
			password, err = readPassword()
			if err != nil {
				return fmt.Errorf("failed to read password: %w", err)
			}
		} else {
			return err
		}
	}

	// Validate password.
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}
	if !crypto.ValidatePassword(password, hash) {
		return errors.New("invalid password")
	}

	if Verbose {
		fmt.Println("Encrypting content...")
	}
	// Encrypt content.
	encrypted, err := crypto.Encrypt(string(content), password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Upload to vault.
	message := fmt.Sprintf("Add/update secret: %s", vaultPath)
	if Verbose {
		fmt.Println("Uploading to vault...")
	}
	if err := client.UploadFile(ctx, vaultPath, []byte(encrypted), message); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	if Verbose {
		fmt.Printf("Uploading secret to vault at path: %s\n", vaultPath)
	}
	fmt.Printf("Secret '%s' uploaded to vault\n", vaultPath)
	return nil
}