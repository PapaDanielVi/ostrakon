package commands

import (
	"errors"
	"fmt"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/crypto"
	"github.com/spf13/cobra"
)

var setGlobalMasterCmd = &cobra.Command{
	Use:   "set-global-master <password>",
	Short: "Set a global master password to avoid repeated prompts",
	Long: `Store your master password in the system keychain for convenience.
When set, the master password will not be prompted for each operation.
Use this only on trusted machines where you control the keychain.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetGlobalMaster,
}

func runSetGlobalMaster(cmd *cobra.Command, args []string) error {
	// Check if initialized
	if _, err := config.GetToken(); err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Get password hash for validation
	hash, err := config.GetPasswordHash()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Get password (either from argument or prompt)
	var password string
	if len(args) > 0 {
		password = args[0]
	} else {
		password, err = readPassword()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	}

	// Validate against stored hash
	if !crypto.ValidatePassword(password, hash) {
		return errors.New("invalid master password")
	}

	// Store in keychain
	if err := config.StoreGlobalMasterPassword(password); err != nil {
		return fmt.Errorf("failed to store global master password: %w", err)
	}

	fmt.Println("Global master password stored successfully")
	fmt.Println("Note: This allows operations without password prompts")
	return nil
}
