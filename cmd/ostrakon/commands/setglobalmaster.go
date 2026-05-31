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
	Short: "Set a global master password (deprecated - now automatic during init)",
	Long: `Store your master password in the system keychain.

NOTE: This command is deprecated. The master password is now automatically
stored in the keyring during 'ostrakon init'. Use --no-keyring with init
to opt out of this behavior.

This command remains for backward compatibility.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSetGlobalMaster,
}

func runSetGlobalMaster(cmd *cobra.Command, args []string) error {
	// Check if initialized
	if _, err := config.GetToken(); err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	fmt.Println("Note: master password is now automatically stored during 'init'.")
	fmt.Println("This command exists for backward compatibility.")

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

	// Store in keyring
	if err := config.StoreGlobalMasterPassword(password); err != nil {
		return fmt.Errorf("failed to store global master password: %w", err)
	}

	fmt.Println("Global master password stored successfully")
	return nil
}
