package commands

import (
	"fmt"
	"os"

	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/provider"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <name> [-o file]",
	Short: "Get and decrypt a secret from the vault",
	Long: `Download and decrypt a secret from the vault.
Outputs to stdout by default or to a file if -o is specified.`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

var (
	outputFile string
	getProfile string
)

func init() {
	getCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: stdout)")
	getCmd.Flags().StringVarP(&getProfile, "profile", "p", "", "Profile/namespace for the file")
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get vault client from provider factory
	client, err := provider.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Determine vault path
	name := args[0]
	vaultPath := name
	if getProfile != "" {
		vaultPath = fmt.Sprintf("profiles/%s/%s", getProfile, name)
	}

	// Download from vault
	encrypted, err := client.DownloadFile(ctx, vaultPath)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Prompt for master password (always prompt for read operations)
	password, err := getPasswordForRead()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	// Decrypt content
	decrypted, err := crypto.Decrypt(string(encrypted), password)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Output
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(decrypted), 0600); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Secret saved to %s\n", outputFile)
	} else {
		fmt.Print(decrypted)
	}

	return nil
}
