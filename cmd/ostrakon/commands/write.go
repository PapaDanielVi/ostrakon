package commands

import (
	"fmt"
	"os"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
)

var writeCmd = &cobra.Command{
	Use:   "write <name> [-o file]",
	Short: "Write a decrypted secret to a file",
	Long: `Download and decrypt a secret from the vault and write it to a file.
The master password is always prompted for security.`,
	Args: cobra.ExactArgs(1),
	RunE:  runWrite,
}

var writeOutputFile string

func init() {
	writeCmd.Flags().StringVarP(&writeOutputFile, "output", "o", "", "Output file (default: uses secret name)")
}

func runWrite(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

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

	// Download from vault
	encrypted, err := client.DownloadFile(ctx, name)
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

	// Determine output filename
	outputPath := writeOutputFile
	if outputPath == "" {
		outputPath = name
	}

	// Ensure parent directory exists (if path contains a slash)
	if idx := findLastSlash(outputPath); idx > 0 {
		dir := outputPath[:idx]
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(decrypted), 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Secret written to %s\n", outputPath)
	return nil
}

// findLastSlash returns the index of the last slash in the path, or -1 if not found.
func findLastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return i
		}
	}
	return -1
}