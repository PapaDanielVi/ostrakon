// Package crypto provides functions for encrypting and decrypting data using AES-256-GCM with keys derived from passwords using Argon2id.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters
	Argon2Time    = 3
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4
	Argon2KeyLen  = 32 // 256 bits for AES-256
)

// DeriveKey derives a 32-byte key from the password using Argon2id
func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)
}

// GenerateSalt generates a random salt for key derivation
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// GenerateRandomBytes generates random bytes for IV/nonce
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return b, nil
}

// Encrypt encrypts data using AES-256-GCM
// Returns: base64(salt + nonce + ciphertext + tag)
func Encrypt(plaintext, password string) (string, error) {
	if plaintext == "" {
		return "", errors.New("plaintext cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Generate salt and derive key
	salt, err := GenerateSalt()
	if err != nil {
		return "", err
	}
	key := DeriveKey([]byte(password), salt)

	// Generate random nonce
	nonce := make([]byte, 12) // GCM nonce size
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Create cipher and encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// Package: salt + nonce + ciphertext
	result := make([]byte, 0, len(salt)+len(nonce)+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts AES-256-GCM encrypted data
// Expects: base64(salt + nonce + ciphertext + tag)
func Decrypt(encryptedB64, password string) (string, error) {
	if encryptedB64 == "" {
		return "", errors.New("encrypted data cannot be empty")
	}
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(encryptedB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(data) < 16+12+16 { // salt(16) + nonce(12) + tag(16)
		return "", errors.New("invalid encrypted data: too short")
	}

	// Extract salt, nonce, ciphertext
	salt := data[:16]
	nonce := data[16:28]
	ciphertext := data[28:]

	// Derive key
	key := DeriveKey([]byte(password), salt)

	// Create cipher and decrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed: wrong password or corrupted data")
	}

	return string(plaintext), nil
}

// HashPassword creates a SHA-256 hash for password validation (not for encryption)
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ValidatePassword checks if the provided password matches the stored hash
func ValidatePassword(password, storedHash string) bool {
	return HashPassword(password) == storedHash
}
