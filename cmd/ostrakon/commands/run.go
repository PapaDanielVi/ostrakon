package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <script> [--env=secret]",
	Short: "Run a script with secrets from the vault",
	Long: `Execute a local script using decrypted secrets as environment variables.
Secrets are fetched from the vault, decrypted in memory, and passed to the script.
They vanish from RAM when the script exits.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runScript,
}

var (
	envSecrets []string
)

func init() {
	runCmd.Flags().StringArrayVarP(&envSecrets, "env", "e", []string{}, "Secret name(s) to inject as environment variables")
}

func runScript(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	scriptPath := args[0]

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return fmt.Errorf("script not found: %s", scriptPath)
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

	// Fetch and decrypt each secret
	envVars := os.Environ()
	for _, secretName := range envSecrets {
		encrypted, err := client.DownloadFile(ctx, secretName)
		if err != nil {
			return fmt.Errorf("failed to download secret '%s': %w", secretName, err)
		}

		decrypted, err := crypto.Decrypt(string(encrypted), password)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret '%s': %w", secretName, err)
		}

		// Parse as key=value
		parts := strings.SplitN(decrypted, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("secret '%s' is not in KEY=VALUE format", secretName)
		}
		envVars = append(envVars, fmt.Sprintf("%s=%s", parts[0], parts[1]))
	}

	// Execute script
	execCmd := exec.Command(scriptPath, args[1:]...)
	execCmd.Env = envVars
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}
