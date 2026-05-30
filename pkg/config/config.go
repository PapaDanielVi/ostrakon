package config

import (
	"errors"
	"os"

	"github.com/zalando/go-keyring"
)

const (
	ServiceName     = "ostrakon"
	PATKey          = "github_pat"
	PasswordHashKey = "password_hash"
)

// StorePAT stores the GitHub Personal Access Token in the system keychain
func StorePAT(pat string) error {
	if pat == "" {
		return errors.New("PAT cannot be empty")
	}
	return keyring.Set(ServiceName, PATKey, pat)
}

// GetPAT retrieves the GitHub Personal Access Token from the system keychain
func GetPAT() (string, error) {
	pat, err := keyring.Get(ServiceName, PATKey)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", errors.New("no PAT found. Run 'ostrakon init' first")
		}
		return "", err
	}
	return pat, nil
}

// DeletePAT removes the GitHub PAT from the system keychain
func DeletePAT() error {
	return keyring.Delete(ServiceName, PATKey)
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
