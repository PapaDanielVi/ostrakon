package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Ostrakon vault",
	Long: `Initialize Ostrakon by:
1. Storing your GitHub access token securely in the OS keychain
2. Setting a master password for encryption
3. Verifying connectivity to your repository`,
	RunE: runInit,
}

// initReader is settable for testing
var initReader = bufio.NewReader(os.Stdin)

func runInit(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Check if already initialized
	if _, err := config.GetToken(); err == nil {
		fmt.Println("Ostrakon is already initialized. Use 'ostrakon shred --all' to reset.")
		return nil
	}

	// Step 1: Prompt for repository URL
	fmt.Print("Repository URL (e.g., https://github.com/owner/repo or owner/repo): ")
	repoURL, err := initReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read repository URL: %w", err)
	}
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return fmt.Errorf("repository URL cannot be empty")
	}

	// Parse the repository URL
	owner, repoName, err := github.ParseRepoURL(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}
	fmt.Printf("  Owner: %s\n", owner)
	fmt.Printf("  Repository: %s\n", repoName)

	// Step 2: Prompt for GitHub access token
	fmt.Print("\nEnter your GitHub Personal Access Token (with repo scope): ")
	token, err := initReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	// Step 3: Verify connectivity and authenticate
	fmt.Println("\n[1/3] Checking repository access...")

	// Create client and check connectivity
	client, err := github.NewClient(token, owner, repoName)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.CheckConnectivity(ctx); err != nil {
		return fmt.Errorf("connectivity check failed: %w", err)
	}
	fmt.Println("  ✓ Repository found and accessible")

	// Step 4: Store credentials
	fmt.Println("\n[2/3] Storing credentials securely...")

	if err := config.StoreToken(token); err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}
	fmt.Println("  ✓ Access token stored in keychain")

	if err := config.StoreRepoInfo(repoURL, owner, repoName); err != nil {
		config.DeleteToken()
		return fmt.Errorf("failed to store repo info: %w", err)
	}
	fmt.Println("  ✓ Repository URL stored")

	// Step 5: Prompt for master password
	fmt.Println("\n[3/3] Setting up encryption...")

	password, err := readPassword()
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	passwordConfirm, err := readPasswordPrompt("  Confirm master password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password != passwordConfirm {
		return fmt.Errorf("passwords do not match")
	}

	// Store password hash for validation
	hash := crypto.HashPassword(password)
	if err := config.StorePasswordHash(hash); err != nil {
		return fmt.Errorf("failed to store password hash: %w", err)
	}
	fmt.Println("  ✓ Master password configured")

	// Success message
	fmt.Println("\n✓ Authentication complete!")
	fmt.Printf("\nYour vault is ready at: %s\n", client.RepoURL())
	fmt.Println("\nNext steps:")
	fmt.Println("  ostrakon add <file>   - Add a secret to the vault")
	fmt.Println("  ostrakon ls           - List all secrets")
	fmt.Println("  ostrakon get <name>   - Get and decrypt a secret")

	return nil
}

// readPassword reads a password from stdin without echoing
func readPassword() (string, error) {
	fmt.Fprint(os.Stderr, "Enter master password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr)
	return string(password), nil
}

// readPasswordPrompt reads a password from stdin with a custom prompt
func readPasswordPrompt(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr)
	return string(password), nil
}

// getPassword retrieves the master password, using the global password if available
func getPassword() (string, error) {
	if config.HasGlobalMasterPassword() {
		return config.GetGlobalMasterPassword()
	}
	return readPassword()
}