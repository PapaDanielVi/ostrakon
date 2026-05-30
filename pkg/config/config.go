// PACKAGE config provides functions to store and retrieve configuration data such as GitHub tokens, repository info, and password hashes using the system keychain for secure storage.
package config

import (
	"errors"
	"os"

	"github.com/zalando/go-keyring"
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

// StoreToken stores the GitHub access token in the system keychain
func StoreToken(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	return keyring.Set(ServiceName, TokenKey, token)
}

// GetToken retrieves the GitHub access token from the system keychain
func GetToken() (string, error) {
	token, err := keyring.Get(ServiceName, TokenKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no token found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return token, nil
}

// DeleteToken removes the GitHub token from the system keychain
func DeleteToken() error {
	return keyring.Delete(ServiceName, TokenKey)
}

// StoreRepoInfo stores the repository URL and parsed owner/repo name
func StoreRepoInfo(url, owner, name string) error {
	if url == "" || owner == "" || name == "" {
		return errors.New("repo URL, owner, and name are required")
	}
	if err := keyring.Set(ServiceName, RepoURLKey, url); err != nil {
		return err
	}
	if err := keyring.Set(ServiceName, RepoOwnerKey, owner); err != nil {
		return err
	}
	return keyring.Set(ServiceName, RepoNameKey, name)
}

// GetRepoURL retrieves the stored repository URL
func GetRepoURL() (string, error) {
	url, err := keyring.Get(ServiceName, RepoURLKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no repo URL found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return url, nil
}

// GetRepoOwner retrieves the stored repository owner
func GetRepoOwner() (string, error) {
	owner, err := keyring.Get(ServiceName, RepoOwnerKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no repo owner found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return owner, nil
}

// GetRepoName retrieves the stored repository name
func GetRepoName() (string, error) {
	name, err := keyring.Get(ServiceName, RepoNameKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no repo name found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return name, nil
}

// DeleteRepoInfo removes all repo info from the keychain
func DeleteRepoInfo() error {
	if err := keyring.Delete(ServiceName, RepoURLKey); err != nil && err != keyring.ErrNotFound {
		return err
	}
	if err := keyring.Delete(ServiceName, RepoOwnerKey); err != nil && err != keyring.ErrNotFound {
		return err
	}
	return keyring.Delete(ServiceName, RepoNameKey)
}

// StorePasswordHash stores the hashed password validation checksum
func StorePasswordHash(hash string) error {
	if hash == "" {
		return errors.New("hash cannot be empty")
	}
	return keyring.Set(ServiceName, PasswordHashKey, hash)
}

// GetPasswordHash retrieves the stored password hash for validation
func GetPasswordHash() (string, error) {
	hash, err := keyring.Get(ServiceName, PasswordHashKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no password hash found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return hash, nil
}

// DeletePasswordHash removes the password hash from the keychain
func DeletePasswordHash() error {
	return keyring.Delete(ServiceName, PasswordHashKey)
}

// ConfigDir returns the user's Ostrakon config directory
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".ostrakon"
	}
	return home + "/.ostrakon"
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	dir := ConfigDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0700)
	}
	return nil
}

// StoreGlobalMasterPassword stores the master password directly in the keyring
func StoreGlobalMasterPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	return keyring.Set(ServiceName, GlobalMasterPasswordKey, password)
}

// GetGlobalMasterPassword retrieves the stored master password from the keyring
func GetGlobalMasterPassword() (string, error) {
	password, err := keyring.Get(ServiceName, GlobalMasterPasswordKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no global master password found. Run 'ostrakon set-global-master <password>' first")
		}
		return "", err
	}
	return password, nil
}

// DeleteGlobalMasterPassword removes the global master password from the keyring
func DeleteGlobalMasterPassword() error {
	return keyring.Delete(ServiceName, GlobalMasterPasswordKey)
}

// HasGlobalMasterPassword checks if a global master password is stored
func HasGlobalMasterPassword() bool {
	_, err := keyring.Get(ServiceName, GlobalMasterPasswordKey)
	return err == nil
}
