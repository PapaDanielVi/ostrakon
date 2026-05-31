package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit a secret in the vault",
	Long: `Download, decrypt, edit, and re-encrypt a secret from the vault.
Uses the $EDITOR environment variable or falls back to 'vim'.`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

var allowedEditors = map[string]struct{}{
	"vim":  {},
	"nvim": {},
	"nano": {},
	"vi":   {},
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	// Get editor
	editor := os.Getenv("EDITOR")
	if len(editor) == 0 {
		editor = "vim"
	}

	editor = filepath.Base(strings.TrimSpace(editor))
	if _, ok := allowedEditors[editor]; !ok {
		return fmt.Errorf("unsupported editor: %q", editor)
	}

	// Write to temp file
	tmpFile, err := os.CreateTemp("", "ostrakon-edit-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.WriteString(decrypted); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	_ = tmpFile.Close()

	// #nosec G702 -- editor is validated against a fixed allowlist above
	editCmd := exec.CommandContext(ctx, editor, tmpPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return fmt.Errorf("editor '%s' not found", editor)
		}
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
	newContent, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	// Get password for re-encryption (use keyring if available)
	reCryptPassword, err := getPassword()
	if err != nil {
		return fmt.Errorf("failed to read password for re-encryption: %w", err)
	}

	// Encrypt and upload
	encryptedStr, err := crypto.Encrypt(string(newContent), reCryptPassword)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	if err := client.UploadFile(ctx, name, []byte(encryptedStr), fmt.Sprintf("Edit secret: %s", name)); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	fmt.Printf("Secret '%s' updated\n", name)
	return nil
}
