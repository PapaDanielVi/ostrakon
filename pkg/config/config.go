// Package config provides functions to store and retrieve configuration data such as GitHub tokens, repository info, and password hashes using the system keychain for secure storage.
package config

import (
	"errors"
	"os"

	"github.com/PapaDanielVi/ostrakon/pkg/keyring"
)

const (
	ServiceName             = "ostrakon"
	TokenKey                = "github_token"
	RepoURLKey              = "repo_url"
	RepoOwnerKey            = "repo_owner"
	RepoNameKey             = "repo_name"
	PasswordHashKey         = "password_hash"
	GlobalMasterPasswordKey = "global_master_password"
)

// keyringClient is the keyring implementation used for storage.
// It defaults to the system keychain but can be overridden for testing.
var keyringClient keyring.Keyring = keyring.DefaultKeyring

// SetKeyring sets a custom keyring implementation (for testing).
func SetKeyring(k keyring.Keyring) {
	keyringClient = k
}

// GetKeyring returns the current keyring implementation.
func GetKeyring() keyring.Keyring {
	return keyringClient
}

// StoreToken stores the GitHub access token in the system keychain.
func StoreToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	return keyringClient.Set(ServiceName, TokenKey, token)
}

// GetToken retrieves the GitHub access token from the system keychain.
func GetToken() (string, error) {
	token, err := keyringClient.Get(ServiceName, TokenKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("no token found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return token, nil
}

// DeleteToken removes the GitHub token from the system keychain.
func DeleteToken() error {
	return keyringClient.Delete(ServiceName, TokenKey)
}

// StoreRepoInfo stores the repository URL and parsed owner/repo name.
func StoreRepoInfo(url, owner, name string) error {
	if url == "" || owner == "" || name == "" {
		return errors.New("repo URL, owner, and name are required")
	}
	if err := keyringClient.Set(ServiceName, RepoURLKey, url); err != nil {
		return err
	}
	if err := keyringClient.Set(ServiceName, RepoOwnerKey, owner); err != nil {
		return err
	}
	return keyringClient.Set(ServiceName, RepoNameKey, name)
}

// GetRepoURL retrieves the stored repository URL.
func GetRepoURL() (string, error) {
	url, err := keyringClient.Get(ServiceName, RepoURLKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("no repo URL found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return url, nil
}

// GetRepoOwner retrieves the stored repository owner.
func GetRepoOwner() (string, error) {
	owner, err := keyringClient.Get(ServiceName, RepoOwnerKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("no repo owner found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return owner, nil
}

// GetRepoName retrieves the stored repository name.
func GetRepoName() (string, error) {
	name, err := keyringClient.Get(ServiceName, RepoNameKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("no repo name found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return name, nil
}

// DeleteRepoInfo removes all repo info from the keychain.
func DeleteRepoInfo() error {
	_ = keyringClient.Delete(ServiceName, RepoURLKey)
	_ = keyringClient.Delete(ServiceName, RepoOwnerKey)
	return keyringClient.Delete(ServiceName, RepoNameKey)
}

// StorePasswordHash stores the hashed password validation checksum.
func StorePasswordHash(hash string) error {
	if hash == "" {
		return errors.New("hash cannot be empty")
	}
	return keyringClient.Set(ServiceName, PasswordHashKey, hash)
}

// GetPasswordHash retrieves the stored password hash for validation.
func GetPasswordHash() (string, error) {
	hash, err := keyringClient.Get(ServiceName, PasswordHashKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("no password hash found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return hash, nil
}

// DeletePasswordHash removes the password hash from the keychain.
func DeletePasswordHash() error {
	return keyringClient.Delete(ServiceName, PasswordHashKey)
}

// ConfigDir returns the user's Ostrakon config directory.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ostrakon"
	}
	return home + "/.ostrakon"
}

// EnsureConfigDir ensures the config directory exists.
func EnsureConfigDir() error {
	dir := ConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0700)
	}
	return nil
}

// StoreGlobalMasterPassword stores the master password directly in the keyring.
func StoreGlobalMasterPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	return keyringClient.Set(ServiceName, GlobalMasterPasswordKey, password)
}

// GetGlobalMasterPassword retrieves the stored master password from the keyring.
func GetGlobalMasterPassword() (string, error) {
	password, err := keyringClient.Get(ServiceName, GlobalMasterPasswordKey)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", errors.New("no global master password found. Run 'ostrakon set-global-master <password>' first")
		}
		return "", err
	}
	return password, nil
}

// DeleteGlobalMasterPassword removes the global master password from the keyring.
func DeleteGlobalMasterPassword() error {
	return keyringClient.Delete(ServiceName, GlobalMasterPasswordKey)
}

// HasGlobalMasterPassword checks if a global master password is stored.
func HasGlobalMasterPassword() bool {
	_, err := keyringClient.Get(ServiceName, GlobalMasterPasswordKey)
	return err == nil
}
