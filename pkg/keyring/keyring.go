// Package keyring provides an interface for secure key-value storage operations.
package keyring

import (
	gokeyring "github.com/zalando/go-keyring"
)

// ErrNotFound is returned when a key is not found in the keyring.
var ErrNotFound = gokeyring.ErrNotFound

// Keyring defines the interface for secure key-value storage operations.
type Keyring interface {
	Set(service, key, value string) error
	Get(service, key string) (string, error)
	Delete(service, key string) error
}

// defaultKeyring wraps the go-keyring library implementation.
type defaultKeyring struct{}

// DefaultKeyring is the production implementation using the system keychain.
var DefaultKeyring Keyring = &defaultKeyring{}

// Set stores a value in the system keyring.
func (k *defaultKeyring) Set(service, key, value string) error {
	return gokeyring.Set(service, key, value)
}

// Get retrieves a value from the system keyring.
func (k *defaultKeyring) Get(service, key string) (string, error) {
	return gokeyring.Get(service, key)
}

// Delete removes a value from the system keyring.
func (k *defaultKeyring) Delete(service, key string) error {
	return gokeyring.Delete(service, key)
}